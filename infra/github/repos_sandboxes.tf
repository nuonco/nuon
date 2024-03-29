module "nuon-sandboxes" {
  source           = "./modules/repository"
  name             = "sandboxes"
  description      = "Builtin sandboxes for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

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

  providers = {
    github = github.nuon
  }
}

module "nuon-azure-aks-byopn-sandbox" {
  source           = "./modules/repository"
  name             = "terraform-azure-aks-byovpn-sandbox"
  description      = "Azure AKS BYOVPN sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}
