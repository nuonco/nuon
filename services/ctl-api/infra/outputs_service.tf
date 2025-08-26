output "tfe_token" {
  value = var.tfe_token
}

output "certificate_arn" {
  value = module.cert.acm_certificate_arn
}

output "public_domain" {
  value = module.cert.domain
}

output "internal_domain" {
  value = module.internal-cert.domain
}

output "internal_certificate_arn" {
  value = module.internal-cert.acm_certificate_arn
}

output "runner_domain" {
  value = module.runner-cert.domain
}

output "runner_certificate_arn" {
  value = module.runner-cert.acm_certificate_arn
}

output "install_templates_bucket" {
  value = local.bucket_name
}

output "install_templates_bucket_base_url" {
  value = "https://${module.bucket.s3_bucket_bucket_regional_domain_name}/"
}

output "install_templates_bucket_region" {
  value = module.bucket.s3_bucket_region
}

// management for working with k8s
output "org_runner" {
  description = "configuration to talk to the runner k8s cluster"
  value = {
    k8s_ca_data         = nonsensitive(data.tfe_outputs.infra-orgs.values.runner_k8s.ca_data)
    k8s_public_endpoint = nonsensitive(data.tfe_outputs.infra-orgs.values.runner_k8s.public_endpoint)
    k8s_cluster_id      = nonsensitive(data.tfe_outputs.infra-orgs.values.runner_k8s.cluster_id)
    k8s_iam_role_arn    = nonsensitive(data.tfe_outputs.infra-orgs.values.runner_k8s.access_role_arns["eks-ctl-api"])
    oidc_provider_arn   = nonsensitive(data.tfe_outputs.infra-orgs.values.runner_k8s.oidc_provider_arn)
    oidc_provider_url   = nonsensitive(data.tfe_outputs.infra-orgs.values.runner_k8s.oidc_provider_url)
    region              = local.vars.region
    support_role_arn    = nonsensitive(data.tfe_outputs.infra-orgs.values.iam_roles.support.arn)
  }
}

output "dns" {
  description = "dns management"

  value = {
    enabled                 = true,
    management_iam_role_arn = module.public_dns_access_role.iam_role_arn,
    zone_id                 = nonsensitive(data.tfe_outputs.infra-orgs.values.public_domain.zone_id),
    root_domain             = nonsensitive(data.tfe_outputs.infra-orgs.values.public_domain.domain),
  }
}

output "management" {
  description = "management account for iam, ecr and more"

  value = {
    iam_role_arn     = module.management_access_role.iam_role_arn,
    account_id       = nonsensitive(data.tfe_outputs.infra-orgs.values.ecr.account_id),
    ecr_registry_id  = nonsensitive(data.tfe_outputs.infra-orgs.values.ecr.registry_id),
    ecr_registry_arn = nonsensitive(data.tfe_outputs.infra-orgs.values.ecr.registry_arn),
  }
}
