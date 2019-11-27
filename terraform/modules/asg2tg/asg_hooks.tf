resource "aws_autoscaling_lifecycle_hook" "node_launch" {
  count                  = length(var.autoscaling_groups_names)
  autoscaling_group_name = var.autoscaling_groups_names[count.index]
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_LAUNCHING"
  name                   = "instance_launch"
  default_result         = "CONTINUE"
  heartbeat_timeout      = 30
}

resource "aws_autoscaling_lifecycle_hook" "node_terminate" {
  count                  = length(var.autoscaling_groups_names)
  autoscaling_group_name = var.autoscaling_groups_names[count.index]
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_TERMINATING"
  name                   = "instance_terminate"
  default_result         = "CONTINUE"
  heartbeat_timeout      = 30
}
