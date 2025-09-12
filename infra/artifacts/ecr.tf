module "nuonctl" {
  source = "../modules/ecr"

  name = "nuonctl"
  tags = {
    artifact      = "nuonctl"
    artifact_type = "binary"
  }

  region = local.aws_settings.region
  providers = {
    aws = aws.infra-shared-prod
  }
}

module "cli" {
  source = "../modules/public-ecr"

  name        = "cli"
  description = "Nuon cli"
  about       = "Nuon cli"
  tags = {
    artifact      = "cli"
    artifact_type = "binary"
  }

  region = local.aws_settings.public_region
  providers = {
    aws = aws.public
  }
}

module "runner" {
  source = "../modules/public-ecr"

  name        = "runner"
  description = "Nuon runner"
  about       = "Nuon runner"
  tags = {
    artifact      = "runner"
    artifact_type = "binary"
  }

  region = local.aws_settings.public_region
  providers = {
    aws = aws.public
  }
}

module "stage-runner" {
  source = "../modules/public-ecr"

  name        = "stage-runner"
  description = "Nuon runner stage"
  about       = "Nuon runner stage"
  tags = {
    artifact      = "stage-runner"
    artifact_type = "binary"
    pre_release   = "true"
  }

  region = local.aws_settings.public_region
  providers = {
    aws = aws.public
  }
}

module "e2e" {
  source = "../modules/public-ecr"

  name = "e2e"
  tags = {
    artifact = "e2e"
  }
  description = "E2E image for testing nuon with an introspection api."
  about       = "E2E image for testing nuon with an introspection api."

  region = local.aws_settings.public_region
  providers = {
    aws = aws.public
  }
}

module "helm_temporal" {
  source = "../modules/public-ecr"

  name        = "helm-temporal"
  region      = local.aws_settings.public_region
  description = "temporal helm chart from mono/charts"
  about       = "Helm chart for installing temporal"
  tags = {
    artifact      = "helm-temporal"
    artifact_type = "helm-oci"
  }

  providers = {
    aws = aws.public
  }
}

module "helm_demo" {
  source = "../modules/public-ecr"

  name        = "helm-demo"
  region      = local.aws_settings.public_region
  description = "Demo helm chart"
  about       = "Demo helm chart"

  tags = {
    artifact      = "helm-demo"
    artifact_type = "helm-oci"
  }

  providers = {
    aws = aws.public
  }
}

module "helm_waypoint" {
  source = "../modules/public-ecr"

  name        = "helm-waypoint"
  region      = local.aws_settings.public_region
  description = "waypoint helm chart from mono/charts"
  about       = "Helm chart for installing waypoint"

  tags = {
    artifact      = "helm-waypoint"
    artifact_type = "helm-oci"
  }

  providers = {
    aws = aws.public
  }
}

module "waypoint_plugin_exp" {
  source = "../modules/public-ecr"

  name        = "waypoint-plugin-exp"
  region      = local.aws_settings.public_region
  description = "nuon waypoint plugin"
  about       = "nuon waypoint plugin"
  tags = {
    artifact      = "waypoint-plugin-exp"
    artifact_type = "waypoint-plugin-odr"
  }

  providers = {
    aws = aws.public
  }
}

module "waypoint_plugin_helm" {
  source = "../modules/public-ecr"

  name        = "waypoint-plugin-helm"
  region      = local.aws_settings.public_region
  description = "nuon waypoint plugin"
  about       = "nuon waypoint plugin"
  tags = {
    artifact      = "waypoint-plugin-helm"
    artifact_type = "waypoint-plugin-odr"
  }

  providers = {
    aws = aws.public
  }
}

module "waypoint_plugin_noop" {
  source = "../modules/public-ecr"

  name        = "waypoint-plugin-noop"
  region      = local.aws_settings.public_region
  description = "nuon waypoint plugin"
  about       = "nuon waypoint plugin"
  tags = {
    artifact      = "waypoint-plugin-noop"
    artifact_type = "waypoint-plugin-odr"
  }

  providers = {
    aws = aws.public
  }
}

module "waypoint_plugin_oci" {
  source = "../modules/public-ecr"

  name        = "waypoint-plugin-oci"
  region      = local.aws_settings.public_region
  description = "nuon waypoint oci plugin"
  about       = "nuon waypoint oci plugin"
  tags = {
    artifact      = "waypoint-plugin-oci"
    artifact_type = "waypoint-plugin-odr"
  }

  providers = {
    aws = aws.public
  }
}

module "waypoint_plugin_oci_sync" {
  source = "../modules/public-ecr"

  name        = "waypoint-plugin-oci-sync"
  region      = local.aws_settings.public_region
  description = "nuon waypoint oci sync plugin"
  about       = "nuon waypoint oci sync plugin"
  tags = {
    artifact      = "waypoint-plugin-oci-sync"
    artifact_type = "waypoint-plugin-odr"
  }

  providers = {
    aws = aws.public
  }
}

module "waypoint_plugin_terraform" {
  source = "../modules/public-ecr"

  name        = "waypoint-plugin-terraform"
  region      = local.aws_settings.public_region
  description = "nuon waypoint plugin"
  about       = "nuon waypoint plugin"
  tags = {
    artifact      = "waypoint-plugin-terraform"
    artifact_type = "waypoint-plugin-odr"
  }

  providers = {
    aws = aws.public
  }
}

module "waypoint_plugin_job" {
  source = "../modules/public-ecr"

  name        = "waypoint-plugin-job"
  region      = local.aws_settings.public_region
  description = "nuon waypoint plugin"
  about       = "nuon waypoint plugin"
  tags = {
    artifact      = "waypoint-plugin-job"
    artifact_type = "waypoint-plugin-odr"
  }

  providers = {
    aws = aws.public
  }
}

module "sandbox_aws_eks" {
  source = "../modules/ecr"

  name   = "sandbox-aws-eks"
  region = local.aws_settings.region
  tags = {
    artifact      = "sandbox-aws-eks"
    artifact_type = "terraform-oci"
  }

  providers = {
    aws = aws.infra-shared-prod
  }
}

module "sandbox_empty" {
  source = "../modules/ecr"

  name   = "sandbox-empty"
  region = local.aws_settings.region
  tags = {
    artifact      = "sandbox-empty"
    artifact_type = "terraform-oci"
  }

  providers = {
    aws = aws.infra-shared-prod
  }
}

module "mirror" {
  source = "../modules/ecr"

  name   = "mirror"
  region = local.aws_settings.region
  tags = {
    artifact      = "mirror"
    artifact_type = "mirrored-repository"
  }

  providers = {
    aws = aws.infra-shared-prod
  }
}

module "docs" {
  source = "../modules/ecr"

  name = "docs"
  tags = {
    artifact      = "docs"
    artifact_type = "binary"
  }

  region = local.aws_settings.region
  providers = {
    aws = aws.infra-shared-prod
  }
}

module "website" {
  source = "../modules/ecr"

  name = "website"
  tags = {
    artifact      = "website"
    artifact_type = "binary"
  }

  region = local.aws_settings.region
  providers = {
    aws = aws.infra-shared-prod
  }
}
