package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"os"
	"strconv"
	"strings"
	"time"
)

type details struct {
	Description string `json:"Description"`
	Details     struct {
		SubnetID         string `json:"Subnet ID"`
		AvailabilityZone string `json:"Availability Zone"`
	} `json:"Details"`
	EndTime              time.Time `json:"EndTime"`
	RequestID            string    `json:"RequestId"`
	ActivityID           string    `json:"ActivityId"`
	Cause                string    `json:"Cause"`
	AutoScalingGroupName string    `json:"AutoScalingGroupName"`
	StartTime            time.Time `json:"StartTime"`
	EC2InstanceID        string    `json:"EC2InstanceId"`
	StatusCode           string    `json:"StatusCode"`
	StatusMessage        string    `json:"StatusMessage"`
}

const (
	detailTypeLaunch    = "EC2 Instance-launch Lifecycle Action"
	detailTypeTerminate = "EC2 Instance-terminate Lifecycle Action"
)

func main() {
	targetGroupARN := os.Getenv("TARGET_GROUP_ARN")
	targetGroupPortStr := os.Getenv("TARGET_GROUP_PORT")
	if targetGroupARN == "" || targetGroupPortStr == "" {
		fmt.Printf("[error] TARGET_GROUP_ARN and TARGET_GROUP_PORT environment variables must be set")
		return
	}
	targetGroupPort, err := strconv.Atoi(targetGroupPortStr)
	if err != nil {
		fmt.Printf("[error] TARGET_GROUP_PORT must be number")
		return
	}

	sess := session.Must(session.NewSession())

	ecSvc := ec2.New(sess)
	asSvc := autoscaling.New(sess)
	elbSvc := elbv2.New(sess)

	lambda.Start(func(ev *events.CloudWatchEvent) error {
		if ev.DetailType != detailTypeLaunch && ev.DetailType != detailTypeTerminate {
			return fmt.Errorf("unknown cloudwatch event detail type %q", ev.DetailType)
		}

		var data details
		if err := json.Unmarshal(ev.Detail, &data); err != nil {
			return fmt.Errorf("unmarshal details: %v", err)
		}
		out, err := ecSvc.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(data.EC2InstanceID)},
		})
		if err != nil {
			return fmt.Errorf("describe instances: %v", err)
		}

		var addresses []string
		for _, reservation := range out.Reservations {
			for _, instance := range reservation.Instances {
				for _, iface := range instance.NetworkInterfaces {
					if len(iface.PrivateIpAddresses) == 0 {
						continue
					}
					for _, address := range iface.PrivateIpAddresses {
						addresses = append(addresses, *address.PrivateIpAddress)
					}
				}
			}
		}

		fmt.Println("PRIVATE ADDRESSES:", strings.Join(addresses, ", "))

		if len(addresses) == 0 {
			return fmt.Errorf("private addresses list is empty")
		}

		port := int64(targetGroupPort)
		var targets []*elbv2.TargetDescription
		for _, addr := range addresses {
			ip := addr // protect from var reusage in loop
			targets = append(targets, &elbv2.TargetDescription{Id: &ip, Port: &port})
		}

		var hookName string
		switch ev.DetailType {
		case detailTypeTerminate:
			hookName = "instance_terminate"
			_, err := elbSvc.DeregisterTargets(&elbv2.DeregisterTargetsInput{
				TargetGroupArn: &targetGroupARN,
				Targets:        targets,
			})
			if err != nil {
				return fmt.Errorf("deregister targets: %s: %v", strings.Join(addresses, ","), err)
			}
		case detailTypeLaunch:
			hookName = "instance_launch"
			_, err := elbSvc.RegisterTargets(&elbv2.RegisterTargetsInput{
				TargetGroupArn: &targetGroupARN,
				Targets:        targets,
			})
			if err != nil {
				return fmt.Errorf("register targets: %s: %v", strings.Join(addresses, ","), err)
			}
		}

		// complete hook
		_, err = asSvc.CompleteLifecycleAction(&autoscaling.CompleteLifecycleActionInput{
			AutoScalingGroupName:  &data.AutoScalingGroupName,
			InstanceId:            &data.EC2InstanceID,
			LifecycleActionResult: aws.String("CONTINUE"),
			LifecycleHookName:     &hookName,
		})
		if err != nil {
			return fmt.Errorf("complete lifecycle action: %v", err)
		}

		return nil
	})
}
