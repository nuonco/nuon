module "lambda_function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "7.7.0"

  function_name  = var.function_name
  image_uri      = var.image_uri
  package_type   = "Image"
  create_package = false

  environment_variables = {
    TABLE = "widgets-${var.install_id}"
  }

  cloudwatch_logs_retention_in_days = 3
  logging_log_group                 = "/aws/lambda/${var.install_id}/${var.function_name}"

  attach_policy_json = true
  policy_json        = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:*"
            ],
            "Resource": ["*"]
      }
  ]
}
EOF
}
