# values for accessing the k8s clusters
output "k8s" {
  value = {
    # roles for individual services that grant access to the k8s cluster
    access_role_arns = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.auth_map_additional_role_arns),

    # information needed to access the k8s cluster
    cluster_id        = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_id),
    ca_data           = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_certificate_authority_data),
    public_endpoint   = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_endpoint),
    oidc_provider_url = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.oidc_provider)
    oidc_provider_arn = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.oidc_provider_arn)
  }
}

output "org_iam_role_name_templates" {
  value = {
    # this is the org iam role that grants access to the deployments bucket with it's prefix
    deployments_access = "arn:aws:iam::${local.org_account_id}:role/orgs/%[1]s/org-deployments-access-%[1]s"
    # this is the org iam role that grants access to the key values bucket and key
    secrets_access = "arn:aws:iam::${local.org_account_id}:role/orgs/%[1]s/org-secrets-access-%[1]s"
    # this is the org iam role that grants access to the installations bucket with it's prefix
    installations_access = "arn:aws:iam::${local.org_account_id}:role/orgs/%[1]s/org-installations-access-%[1]s"
    # this is the org specific role that grants the instances workflow access
    instances_access = "arn:aws:iam::${local.org_account_id}:role/orgs/%[1]s/org-instances-access-%[1]s"
    # this is the org specific role that grants access to the orgs bucket
    orgs_access = "arn:aws:iam::${local.org_account_id}:role/orgs/%[1]s/org-orgs-access-%[1]s"
    # this is the org specific IAM role that is attached to the ODR in our account
    odr = "arn:aws:iam::${local.org_account_id}:role/orgs/%[1]s/org-odr-%[1]s"
    # this is the org specific installer IAM role that workers-installs uses when creating a sandbox
    installer = "arn:aws:iam::${local.org_account_id}:role/orgs/%[1]s/org-installer-%[1]s"
  }
}

output "api" {
  value = {
    url = nonsensitive("orgs-api.${data.tfe_outputs.infra-eks.values.private_zone}"),
  }
}

# outputs for working with the org's account ECR registry
output "ecr" {
  value = {
    registry_arn = "arn:aws:ecr:${local.region}:${local.org_account_id}:repository",
    region       = local.region
    registry_id  = local.org_account_id
  }
}

output "waypoint" {
  value = {
    root_domain            = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.root_domain)
    token_secret_namespace = "default"
    token_secret_template  = "waypoint-bootstrap-token-%[1]s"

    // waypoint servers and runners live in the org account+cluster.
    cluster_id        = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_id),
    ca_data           = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_certificate_authority_data),
    public_endpoint   = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_endpoint),
    oidc_provider_url = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.oidc_provider)
    oidc_provider_arn = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.oidc_provider_arn)
  }
}

output "bootstrap_waypoint" {
  value = {
    domain                 = "bootstrap.${nonsensitive(data.tfe_outputs.infra-eks-orgs.values.root_domain)}"
    token_secret_namespace = "waypoint"
    token_secret_template  = "waypoint-server-token"

    // bootstrap waypoint runs in the orgs account+cluster.
    cluster_id        = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_id),
    ca_data           = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_certificate_authority_data),
    public_endpoint   = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_endpoint),
    oidc_provider_url = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.oidc_provider)
    oidc_provider_arn = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.oidc_provider_arn)
  }
}


# buckets for storing/managing state related to an org
output "buckets" {
  value = {
    deployments = {
      name   = module.deployments_bucket.s3_bucket_id
      region = module.deployments_bucket.s3_bucket_region
    }
    secrets = {
      name   = module.secrets_bucket.s3_bucket_id
      region = module.secrets_bucket.s3_bucket_region
    }
    installations = {
      name   = module.org_installations_bucket.s3_bucket_id
      region = module.org_installations_bucket.s3_bucket_region
    }
    orgs = {
      name   = module.orgs_bucket.s3_bucket_id
      region = module.orgs_bucket.s3_bucket_region
    }
  }
}

output "iam_roles" {
  value = {
    # TODO(jm): should this be scoped per org, instead of a single role?
    install_k8s_access = {
      description = "k8s iam role that allows us to provision infra in a sandbox EKS cluster"
      arn         = module.install_k8s_role_external.iam_role_arn
    }

    support = {
      description = "IAM role that can be assumed, which allows assuming org IAM roles"
      arn         = module.support_role.iam_role_arn
    }
  }
}

output "sandbox" {
  value = {
    bucket = nonsensitive(data.tfe_outputs.sandboxes.values.bucket)
    key    = nonsensitive(data.tfe_outputs.sandboxes.values.key)
  }
}

output "account" {
  value = {
    id = local.org_account_id
  }
}

# the public domain is used for creating domains for installs, within an org.
output "public_domain" {
  value = {
    nameservers = data.aws_route53_zone.public_domain.name_servers
    domain      = data.aws_route53_zone.public_domain.name
    zone_id     = data.aws_route53_zone.public_domain.id
  }
}
