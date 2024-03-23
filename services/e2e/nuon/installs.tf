resource "nuon_install" "main" {
  count  = var.install_count
  app_id = nuon_app.main.id

  name         = "${var.install_prefix}${count.index}"

  dynamic "input" {
    for_each = var.inputs
    content {
      name  = input.value.name
      value = input.value.value
    }
  }

  dynamic "aws" {
    for_each = var.aws
    content {
      iam_role_arn = aws.value.iam_role_arn
      region = aws.value.regions[count.index]
    }
  }

  dynamic "azure" {
    for_each = var.azure
    content {
      location = azure.value.locations[count.index]
      subscription_id = azure.value.subscription_id
      subscription_tenant_id = azure.value.subscription_tenant_id
      service_principal_app_id = azure.value.service_principal_app_id
      service_principal_password = azure.value.service_principal_password
    }
  }

  depends_on = [
    nuon_app_sandbox.main,
    nuon_app_runner.main,
    nuon_job_component.e2e,
    nuon_helm_chart_component.e2e,
  ]
}
