locals {
  honeyhive_customer_name = "honeyhive"
  honehyive_sandbox_id    = nuon_app.sandbox[local.honeyhive_customer_name].id
}

resource "nuon_install" "honeyhive_install" {
  provider = nuon.sandbox

  app_id = local.honehyive_sandbox_id

  name         = "${local.honeyhive_customer_name}-demo"
  region       = "us-east-1"
  iam_role_arn = "customer-${local.honeyhive_customer_name}"

  depends_on = [
    nuon_app_sandbox.sandbox
  ]
}

