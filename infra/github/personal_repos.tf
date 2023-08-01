// most experimentation should happen in the mono repo.
module "code-jonmorehouse" {
  source = "./modules/repository"

  name                     = "code-jonmorehouse"
  description              = "personal workspace for @jonmorehouse"
  enable_ecr               = false
  enable_prod_environment  = false
  enable_stage_environment = false

  topics = ["personal-workspace", "archived"]
}

module "code-jordanacosta" {
  source = "./modules/repository"

  name                     = "code-jordanacosta"
  description              = "personal workspace for @jordanacosta"
  enable_ecr               = false
  enable_prod_environment  = false
  enable_stage_environment = false

  topics = ["personal-workspace"]
}

module "code-focusaurus" {
  source = "./modules/repository"

  name                     = "code-focusaurus"
  description              = "personal workspace for @focusaurus"
  enable_ecr               = false
  enable_prod_environment  = false
  enable_stage_environment = false

  required_checks          = []
  topics = ["personal-workspace"]
}
