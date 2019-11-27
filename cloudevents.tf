resource "aws_cloudwatch_event_rule" "rule" {
  name          = "${local.dashed_name}-${var.name}"
  event_pattern = <<EOF
{
  "source": [
    "aws.autoscaling"
  ],
  "detail-type": [
    "EC2 Instance-launch Lifecycle Action",
    "EC2 Instance-terminate Lifecycle Action"
  ],
  "detail": {
    "AutoScalingGroupName": ${jsonencode(var.autoscaling_groups_names)}
  }
}
EOF
  tags          = local.tags
}

resource "aws_cloudwatch_event_target" "lambda" {
  arn  = aws_lambda_function.lambda.arn
  rule = aws_cloudwatch_event_rule.rule.name
}
resource "aws_lambda_permission" "event_rule" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.rule.arn
}
