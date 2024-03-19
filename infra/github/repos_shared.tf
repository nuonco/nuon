module "shared-warpstream" {
  source = "./modules/repository"

  name            = "warpstream"
  description     = "Nuon configuration for Warpstream."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id
  is_private = true
  enable_branch_protection = false

  collaborators = {
    "caleb-warpstream" = "push"
    "richardartoul" = "push"
    "ryanworl" = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-okteto" {
  source = "./modules/repository"

  name            = "okteto"
  description     = "Nuon configuration for Okteto."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id
  is_private = true
  enable_branch_protection = false

  collaborators = {
      rberrelleza = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-meroxa" {
  source = "./modules/repository"

  name            = "meroxa"
  description     = "Nuon configuration + demo for Meroxa."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id
  is_private = true
  enable_branch_protection = false

  collaborators = {
      simonl2002 = "push"
  }

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
  is_private = true
  enable_branch_protection = false

  collaborators = {
      aduis = "push"
  }

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
  is_private = true
  enable_branch_protection = false

  collaborators = {
      georgemac = "push"
      markphelps = "push"
  }

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
  enable_branch_protection = false
  is_private = true

  collaborators = {
      "michael-hhai" = "push"
      "codehruv" = "push"
  }

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
  is_private = true
  enable_branch_protection = false

  collaborators = {
      thruflo = "push"
      samwillis = "push"
      alco = "push"
      balegas = "push"
  }

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
  is_private = true
  enable_branch_protection = false

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-commonfate" {
  source = "./modules/repository"

  name            = "common-fate"
  description     = "Nuon configuration for common-fate."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id
  is_private = true
  enable_branch_protection = false

  collaborators = {
    chrnorm = "push"
    shwethaumashanker = "push"
    JoshuaWilkes = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-streamkap" {
  source = "./modules/repository"

  name            = "streamkap"
  description     = "Nuon configuration for streamkap."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id
  is_private = true
  enable_branch_protection = false

  collaborators = {
    thomasr888 = "push"
    quang-streamkap = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-100xdev" {
  source = "./modules/repository"

  name            = "100xdev"
  description     = "Nuon configuration for 100xdev."
  required_checks = []
  owning_team_id  = github_team.nuonco-shared.id
  is_private = true
  enable_branch_protection = false

  collaborators = {
    anandsainath = "push"
    shaumik100x = "push"
    sraibagiwith100x = "push"
    gyx119 = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}
