
module "api_gateway" {
  source  = "terraform-aws-modules/apigateway-v2/aws"
  version = "~> 5.3.0"

  name                        = var.name
  hosted_zone_name            = var.domain_name
  domain_name                 = "${var.name}.${var.domain_name}"
  domain_name_certificate_arn = var.domain_name_certificate_arn
  protocol_type               = "HTTP"

  stage_access_log_settings = {
    create_log_group = true
    log_group_name   = "${var.name}-${var.install_id}"
  }

  stage_default_route_settings = {
    detailed_metrics_enabled = true
    throttling_burst_limit   = 100
    throttling_rate_limit    = 100
  }

  # Routes & Integration(s)
  routes = {
    "GET /widgets/{id}" = {
      integration = {
        uri                    = var.lambda_function_arn
        payload_format_version = "2.0"
        timeout_milliseconds   = 12000
      }
    }

    "POST /widgets" = {
      integration = {
        uri                    = var.lambda_function_arn
        payload_format_version = "2.0"
        timeout_milliseconds   = 12000
      }
    }
  }

  cors_configuration = {
    allow_headers = ["content-type", "x-amz-date", "authorization", "x-api-key", "x-amz-security-token", "x-amz-user-agent"]
    allow_methods = ["*"]
    allow_origins = ["*"]
  }
}

resource "aws_lambda_permission" "permission" {
  function_name = element(split(":", var.lambda_function_arn), length(split(":", var.lambda_function_arn)) - 1)
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${module.api_gateway.api_execution_arn}/*/*"
}
