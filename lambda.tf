data "archive_file" "lambda" {
  output_path = "lambda_payload.zip"
  type        = "zip"
  source_file = "${path.module}/files/entrypoint"
}

resource "aws_lambda_function" "lambda" {
  function_name    = "${local.dashed_name}-${var.name}-handler"
  handler          = "entrypoint"
  role             = aws_iam_role.lambda.arn
  runtime          = "go1.x"
  filename         = data.archive_file.lambda.output_path
  source_code_hash = data.archive_file.lambda.output_base64sha256
  environment {
    variables = {
      TARGET_GROUP_ARN  = var.target_group_arn
      TARGET_GROUP_PORT = var.target_group_port
    }
  }
  tags = local.tags
}