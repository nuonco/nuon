locals {
  // TODO(jm): move these to a variable set or other
  subscription_id = "aaf93888-61e7-499e-afa3-a34d780b98a9"
  tenant_id = "c73796d7-1c01-4b07-b625-1815eb63712a"
  client_id = "88c36934-f2e2-43f0-ad42-0eb25f0c5814"
  client_secret = "mhN8Q~nTtzl6_hvNdAvhRtc-Kj.iKEx-3JQrycWj"
}

module "azure-aks" {
  source = "./e2e"

  app_name = "${local.name}-azure-aks"

  sandbox_repo = local.sandboxes_repo
  sandbox_branch = "main"
  sandbox_dir = "azure-aks"
  app_runner_type = "azure-aks"

  install_count = 1
  install_prefix = "azure-aks-"
  azure = [
    {
      locations = ["eastus2"]

      service_principal_password = local.client_secret
      service_principal_app_id = local.client_id
      subscription_id = local.subscription_id
      subscription_tenant_id = local.tenant_id
    }
  ]
}
