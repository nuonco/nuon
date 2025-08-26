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

module "nuon-aws-eks-sandbox-m1" {
  source           = "./modules/repository"
  name             = "aws-eks-sandbox"
  description      = "AWS EKS sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-aws-eks-karpenter-sandbox-m1" {
  source           = "./modules/repository"
  name             = "aws-eks-karpenter-sandbox"
  description      = "AWS EKS + Karpenter sandbox for Nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}
