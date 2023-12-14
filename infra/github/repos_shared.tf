module "shared-meroxa" {
  source = "./modules/repository"

  name            = "meroxa"
  description     = "Nuon configuration + demo for Meroxa."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-weaviate" {
  source = "./modules/repository"

  name            = "weaviate"
  description     = "Nuon configuration + demo for Weaviate."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-flipt" {
  source = "./modules/repository"

  name            = "flipt"
  description     = "Nuon configuration + demo for Flipt."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-honeyhive" {
  source = "./modules/repository"

  name            = "honeyhive"
  description     = "Nuon configuration + demo for Honeyhive."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-electric-sql" {
  source = "./modules/repository"

  name            = "electric-sql"
  description     = "Nuon configuration + demo for Electric SQL."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-clickhouse" {
  source = "./modules/repository"

  name            = "clickhouse"
  description     = "Nuon configuration for clickhouse."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id

  providers = {
    github = github.nuonco-shared
  }
}
