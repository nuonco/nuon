module "nuonctl" {
  source = "../modules/ecr"

  name = "nuonctl"
  tags = {
    artifact      = "sandbox-aws-eks"
    artifact_type = "binary"
  }

  providers = {
    aws = aws.infra-shared-prod
  }
}

module "helm_temporal" {
  source = "../modules/public-ecr"

  name        = "helm-temporal"
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

module "helm_waypoint" {
  source = "../modules/public-ecr"

  name        = "helm-waypoint"
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

module "waypoint_plugin_terraform" {
  source = "../modules/public-ecr"

  name        = "waypoint-plugin-terraform"
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

module "sandbox_aws_eks" {
  source = "../modules/ecr"

  name = "sandbox-aws-eks"
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

  name = "sandbox-empty"
  tags = {
    artifact      = "sandbox-empty"
    artifact_type = "terraform-oci"
  }

  providers = {
    aws = aws.infra-shared-prod
  }
}
