output "deployments_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.deployments.name)
}

output "orgs_deployments_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.deployments_access)
}

output "waypoint_server" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.waypoint)
}

output "orgs_k8s" {
  sensitive = true
  value = {
    ca_data         = data.tfe_outputs.infra-orgs.values.k8s.ca_data
    public_endpoint = data.tfe_outputs.infra-orgs.values.k8s.public_endpoint
    cluster_id      = data.tfe_outputs.infra-orgs.values.k8s.cluster_id
    role_arn        = data.tfe_outputs.infra-orgs.values.k8s.access_role_arns["eks-workers-instances"]
  }
}
