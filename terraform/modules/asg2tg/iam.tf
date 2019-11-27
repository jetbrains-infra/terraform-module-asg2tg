### LAMBDA
data "aws_iam_policy_document" "lambda" {
  statement {
    effect = "Allow"
    actions = [
      "ec2:DescribeInstances",
      "ec2:DescribeNetworkInterfaces"
    ]
    resources = ["*"]
  }

  dynamic "statement" {
    for_each = var.autoscaling_groups_names
    content {
      effect  = "Allow"
      actions = ["autoscaling:CompleteLifecycleAction"]
      resources = [
        "arn:aws:autoscaling:*:*:autoScalingGroup:*:autoScalingGroupName/${statement.value}"
      ]
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "elasticloadbalancing:RegisterTargets",
      "elasticloadbalancing:DeregisterTargets"
    ]
    resources = [var.target_group_arn]
  }
}

resource "aws_iam_policy" "lambda" {
  name   = "${local.dashed_name}-${var.name}"
  policy = data.aws_iam_policy_document.lambda.json
}

resource "aws_iam_role" "lambda" {
  name               = "${local.dashed_name}-${var.name}-lambda"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "lambda.amazonaws.com"
            },
            "Effect": "Allow"
        }
    ]
}
EOF
  tags               = local.tags
}
resource "aws_iam_role_policy_attachment" "lambda" {
  policy_arn = aws_iam_policy.lambda.arn
  role       = aws_iam_role.lambda.id
}
resource "aws_iam_role_policy_attachment" "attach_lambda_policy" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda.id
}

