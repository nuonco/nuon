module "shared_configs" {
  source = "./modules/repository"

  archived    = true
  name        = "shared-configs"
  description = "shared configuration files"
  topics      = ["archived"]
}

module "sandboxes" {
  source = "./modules/repository"

  archived    = true
  name        = "sandboxes"
  description = "terraform modules for sandbox creation"
  topics      = ["terraform", "archived"]
}

module "apks" {
  source = "./modules/repository"

  archived    = true
  name        = "apks"
  description = "repo for building apks used in our images"
  topics      = ["terraform", "archived"]
}

module "chart-common" {
  source = "./modules/repository"

  archived    = true
  name        = "chart-common"
  description = "repo for common charts"
  topics      = ["terraform", "helm", "archived"]
}

module "ci-images" {
  source = "./modules/repository"

  archived    = true
  name        = "ci-images"
  description = "repo for ci specific container images"
  topics      = ["terraform", "archived"]
}

module "action-app-token" {
  source = "./modules/repository"

  name        = "action-app-token"
  description = "shared github token action"
  topics      = ["github-actions", "archived"]
  archived    = true
}

module "action-pr-checks" {
  source = "./modules/repository"

  name        = "action-pr-checks"
  description = "shared action for checking PRs"
  topics      = ["github-actions", "archived"]
  archived    = true
}

module "action-reviewdog" {
  source = "./modules/repository"

  name        = "action-reviewdog"
  description = "repo for shared action linting"
  topics      = ["github-actions", "archived"]
  archived    = true
}

module "action-setup-ci" {
  source = "./modules/repository"

  name        = "action-setup-ci"
  description = "github action to set up ci"
  topics      = ["github-actions", "archived"]
  archived    = true
}

module "action-setup-helm" {
  source = "./modules/repository"

  name        = "action-setup-helm"
  description = "github action to set up helm"
  topics      = ["github-actions", "archived"]
  archived    = true
}

module "action-setup-node" {
  source = "./modules/repository"

  name        = "action-setup-node"
  description = "github action to set up node"
  topics      = ["github-actions", "archived"]
  archived    = true
}

module "action-tf-output" {
  source = "./modules/repository"

  name        = "action-tf-output"
  description = "repo for terraform outputs"
  topics      = ["github-actions", "archived"]
  archived    = true
}
module "api" {
  source = "./modules/repository"

  name        = "api"
  description = "repo for nuon grpc api"
  topics      = ["terraform", "helm", "archived"]
  enable_ecr  = true
  archived    = true

  enable_stage_environment = true
  enable_prod_environment  = true
}

module "api-gateway" {
  source = "./modules/repository"

  name        = "api-gateway"
  description = "repo for the graphql api gateway"
  enable_ecr  = true
  topics      = ["terraform", "helm", "archived"]
  archived    = true
}

module "awesome-customer-cloud" {
  source = "./modules/repository"

  archived    = true
  name        = "awesome-customer-cloud"
  description = "experimental customer cloud repo"
  topics      = ["experimental", "archived"]
}

module "go-aws-assume-role" {
  source = "./modules/repository"

  archived    = true
  name        = "go-aws-assume-role"
  description = "shared tooling for assuming IAM roles"
  topics      = ["archived"]
}

module "go-common" {
  source = "./modules/repository"

  archived    = true
  name        = "go-common"
  description = "repo common shared golang"
  topics      = ["go-lib", "archived"]
}

module "go-components" {
  source = "./modules/repository"

  archived    = true
  name        = "go-components"
  description = "repo for shared component configurations"
  topics      = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-config" {
  source = "./modules/repository"

  archived    = true
  name        = "go-config"
  description = "repo for go service config"
  topics      = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-fetch" {
  source = "./modules/repository"

  archived    = true
  name        = "go-fetch"
  description = "repo for fetching"
  topics      = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-generics" {
  source = "./modules/repository"

  archived    = true
  name        = "go-generics"
  description = "go package for shared generic functions"

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-helm" {
  source = "./modules/repository"

  archived    = true
  name        = "go-helm"
  description = "go package for shared helm tooling"

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-kube" {
  source = "./modules/repository"

  archived    = true
  name        = "go-kube"
  description = "go package for shared kubernetes tooling"

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-sender" {
  source = "./modules/repository"

  archived    = true
  name        = "go-sender"
  description = "go package for shared notification tooling"

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-terraform" {
  source = "./modules/repository"

  archived    = true
  name        = "go-terraform"
  description = "go package for shared terraform tooling"

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-shared-types" {
  source = "./modules/repository"

  archived    = true
  name        = "go-shared-types"
  description = "HACK: go package for shared types"

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-uploader" {
  source = "./modules/repository"

  archived    = true
  name        = "go-uploader"
  description = "go package for upload to s3 tooling"

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-waypoint" {
  source = "./modules/repository"

  archived    = true
  name        = "go-waypoint"
  description = "go package for shared waypoint tooling"

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "go-workflows-meta" {
  source = "./modules/repository"

  archived    = true
  name        = "go-workflows-meta"
  description = "go package for shared tooling for writing out workflow related metadata."

  topics = ["go-lib", "from-template-go-lib", "archived"]
}

module "graphql-api" {
  source = "./modules/repository"

  archived    = true
  name        = "graphql-api"
  description = "repo for nuon graphql api"
  topics      = ["terraform", "helm", "archived"]
  enable_ecr  = true

  enable_stage_environment = true
  enable_prod_environment  = true
}

module "nuonctl" {
  source = "./modules/repository"

  archived    = true
  name        = "nuonctl"
  description = "an experimental cli with automations and easy ways to interact with the api"

  topics = ["experimental", "archived"]
}

module "orgs-api" {
  source = "./modules/repository"

  archived    = true
  name        = "orgs-api"
  description = "repo for nuon orgs api, which exposes infrastructure details of orgs, installs and deployments."
  topics      = ["terraform", "helm", "archived"]
  enable_ecr  = true

  enable_stage_environment = true
  enable_prod_environment  = true
}

module "protos" {
  source = "./modules/repository"

  archived    = true
  name        = "protos"
  description = "mono repo of protocol buffers that power apis, workflows and internal systems - using buf.build."
  topics      = ["archived"]
}

module "template-go-library" {
  source = "./modules/repository"

  archived    = true
  name        = "template-go-library"
  description = "Template for creating a new go library."
  topics      = ["template", "archived"]
}

module "template-go-service" {
  source = "./modules/repository"

  archived                 = true
  name                     = "template-go-service"
  description              = "Template for creating a go service repository."
  enable_ecr               = true
  enable_prod_environment  = true
  enable_stage_environment = true

  topics = ["template", "archived"]
}

module "workers-apps" {
  source = "./modules/repository"

  archived                 = true
  name                     = "workers-apps"
  description              = "temporal workers for apps"
  enable_ecr               = true
  enable_prod_environment  = true
  enable_stage_environment = true

  topics = ["helm", "terraform", "from-template-go-service", "archived"]
}

module "workers-deployments" {
  source = "./modules/repository"

  archived                 = true
  name                     = "workers-deployments"
  description              = "temporal workers for deployments"
  enable_ecr               = true
  enable_prod_environment  = true
  enable_stage_environment = true

  topics = ["helm", "terraform", "from-template-go-service", "archived"]
}

module "workers-executors" {
  source = "./modules/repository"

  archived                 = true
  name                     = "workers-executors"
  description              = "temporal workers for managing infrastructure"
  enable_ecr               = true
  enable_prod_environment  = true
  enable_stage_environment = true

  topics = ["helm", "terraform", "from-template-go-service", "archived"]
}

module "workers-installs" {
  source = "./modules/repository"

  archived                 = true
  name                     = "workers-installs"
  description              = "temporal workers for installs"
  enable_ecr               = true
  enable_prod_environment  = true
  enable_stage_environment = true

  topics = ["helm", "terraform", "from-template-go-service", "archived"]
}

module "workers-instances" {
  source = "./modules/repository"

  archived                 = true
  name                     = "workers-instances"
  description              = "temporal workers for instances"
  enable_ecr               = true
  enable_prod_environment  = true
  enable_stage_environment = true

  topics = ["helm", "terraform", "from-template-go-service", "archived"]
}

module "workers-orgs" {
  source = "./modules/repository"

  archived                 = true
  name                     = "workers-orgs"
  description              = "temporal workers for orgs"
  enable_ecr               = true
  enable_prod_environment  = true
  enable_stage_environment = true

  topics = ["helm", "terraform", "from-template-go-service", "archived"]
}

module "infra-aws" {
  source = "./modules/repository"

  name        = "infra-aws"
  archived    = true
  description = "terraform module for managing the aws org and accounts"
  topics      = ["terraform", "archived"]
}

module "infra-eks-nuon" {
  source = "./modules/repository"

  archived    = true
  name        = "infra-eks-nuon"
  description = "terraform module for managing EKS"
  topics      = ["terraform", "archived"]
}

module "infra-github" {
  source = "./modules/repository"

  archived    = true
  name        = "infra-github"
  description = "terraform module for managing github"
  topics      = ["terraform", "archived"]
}

module "infra-grafana" {
  source = "./modules/repository"

  archived    = true
  name        = "infra-grafana"
  description = "terraform module for managing grafana cloud"
  topics      = ["terraform", "archived"]
}

module "infra-nuon-dns" {
  source = "./modules/repository"

  name        = "infra-nuon-dns"
  archived    = true
  description = "terraform module for managing the nuon.co domain"
  topics      = ["terraform", "archived"]
}

module "infra-orgs" {
  source = "./modules/repository"

  archived    = true
  name        = "infra-orgs"
  description = "terraform module for managing org resources, such as installations, runs and builds."
  topics      = ["terraform", "archived"]
}

module "infra-powertools" {
  source = "./modules/repository"

  name        = "infra-powertools"
  archived    = true
  description = "terraform module for managing the powertools.dev domain"
  topics      = ["terraform", "archived"]
}

module "infra-temporal" {
  source = "./modules/repository"

  name        = "infra-temporal"
  archived    = true
  description = "terraform module for managing our temporal installations"
  topics      = ["terraform", "helm", "archived"]
  enable_ecr  = true
}

module "infra-terraform" {
  source = "./modules/repository"

  name        = "infra-terraform"
  description = "terraform module for managing our terraform workspaces"
  topics      = ["terraform", "archived"]
}

module "horizon" {
  source = "./modules/repository"

  name        = "horizon"
  description = "repo for managing our horizon based url service"
  topics      = ["terraform"]
  archived    = true

  extra_ecr_repos = ["hashicorp-horizon", "hashicorp-waypoint-hzn"]
}

module "public-docs" {
  source = "./modules/repository"

  name        = "public-docs"
  description = "public documentation"
  topics      = ["archived"]
  archived    = true

  enable_branch_protection = false
}

module "demo" {
  source = "./modules/repository"

  name            = "demo"
  enable_ecr      = false
  description     = "Demo repo for Nuon."
  required_checks = []
  topics          = ["archived"]
  archived        = true
}

module "ui" {
  source = "./modules/repository"

  name        = "ui"
  description = "github repo for our ui"
  topics      = ["archived"]
  archived    = true
}

module "eslint-config-nuon" {
  source = "./modules/repository"

  name        = "eslint-config-nuon"
  description = "eslint config for typescript projects"
  topics      = ["archived"]
  archived    = true
}

module "waypoint" {
  source = "./modules/repository"

  name        = "waypoint"
  description = "Our internal fork of hashicorp/waypoint."
  topics      = ["archived"]
  archived    = "true"
}

module "nuon-azure-aks-byopn-sandbox" {
  source           = "./modules/repository"
  name             = "terraform-azure-aks-byovpn-sandbox"
  description      = "Azure AKS BYOVPN sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"
  archived         = "true"

  providers = {
    github = github.nuon
  }
}

module "nuon-terraform-installer-ui-hosted" {
  source           = "./modules/repository"
  name             = "installer-hosted"
  description      = "Multitenant Hosted Installer UI"
  required_checks  = []
  is_public        = false
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  topics   = ["archived"]
  archived = true

  providers = {
    github = github.nuon
  }
}

module "nuon-aws-eks-sandbox" {
  source           = "./modules/repository"
  name             = "terraform-aws-eks-sandbox"
  description      = "AWS EKS sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  topics   = ["archived"]
  archived = true

  providers = {
    github = github.nuon
  }
}

module "nuon-aws-eks-byovpc-sandbox" {
  source           = "./modules/repository"
  name             = "terraform-aws-eks-byovpc-sandbox"
  description      = "AWS EKS BYOVPC sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  topics   = ["archived"]
  archived = true

  providers = {
    github = github.nuon
  }
}

module "nuon-aws-ecs-sandbox" {
  source           = "./modules/repository"
  name             = "terraform-aws-ecs-sandbox"
  description      = "AWS ECS sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  topics   = ["archived"]
  archived = true

  providers = {
    github = github.nuon
  }
}

module "nuon-aws-ecs-byovpc-sandbox" {
  source           = "./modules/repository"
  name             = "terraform-aws-ecs-byovpc-sandbox"
  description      = "AWS ECS BYOVPC sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  topics   = ["archived"]
  archived = true

  providers = {
    github = github.nuon
  }
}

module "nuon-azure-aks-sandbox" {
  source           = "./modules/repository"
  name             = "terraform-azure-aks-sandbox"
  description      = "Azure AKS sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  topics   = ["archived"]
  archived = true

  providers = {
    github = github.nuon
  }
}
