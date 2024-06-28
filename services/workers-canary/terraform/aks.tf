module "azure-aks" {
  source = "./e2e"

  app_name = "${local.name}-azure-aks"
  create_components = true

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = "main"
  sandbox_dir = "azure-aks"
  app_runner_type = "azure-aks"

  install_count = 1
  install_prefix = "azure-aks-"
  azure = [
    {
      locations = ["eastus2"]

      service_principal_password = var.azure_aks_client_secret
      service_principal_app_id = var.azure_aks_client_id
      subscription_id = var.azure_aks_subscription_id
      subscription_tenant_id = var.azure_aks_tenant_id
    }
  ]
}
