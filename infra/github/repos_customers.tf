module "customer-meroxa" {
  source = "./modules/repository"

  name            = "customer-meroxa"
  description     = "Nuon configuration + demo for Meroxa."
  required_checks = []
  owning_team_id  = github_team.nuon.id

  providers = {
    github = github.nuon
  }
}

module "customer-weaviate" {
  source = "./modules/repository"

  name            = "customer-weaviate"
  description     = "Nuon configuration + demo for Weaviate."
  required_checks = []
  owning_team_id  = github_team.nuon.id

  providers = {
    github = github.nuon
  }
}

module "customer-flipt" {
  source = "./modules/repository"

  name            = "customer-flipt"
  description     = "Nuon configuration + demo for Flipt."
  required_checks = []
  owning_team_id  = github_team.nuon.id

  providers = {
    github = github.nuon
  }
}

module "customer-honeyhive" {
  source = "./modules/repository"

  name            = "customer-honeyhive"
  description     = "Nuon configuration + demo for Honeyhive."
  required_checks = []
  owning_team_id  = github_team.nuon.id

  providers = {
    github = github.nuon
  }
}
