module "shared-paradedb" {
  source = "./modules/repository"

  name                     = "paradedb"
  description              = "Nuon configuration for paradedb."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    philippemnoel = "push"
    sardination   = "push"
    rebasedming   = "push"
    neilyio       = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-parededb-installer" {
  source = "./modules/repository"

  name                     = "paradedb-installer"
  description              = "Installer for paradedb."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {
    philippemnoel = "push"
    sardination   = "push"
    rebasedming   = "push"
    neilyio       = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-carbon-installer" {
  source = "./modules/repository"

  name                     = "carbon-installer"
  description              = "Installer for carbon."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {}

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-athena" {
  source = "./modules/repository"

  name                     = "athena"
  description              = "Nuon configuration for athena."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    "bgeils" = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-athena-installer" {
  source = "./modules/repository"

  name                     = "athena-installer"
  description              = "Installer for Athena."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {
    "bgeils" = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-turntable" {
  source = "./modules/repository"

  name                     = "turntable"
  description              = "Nuon configuration for turntable."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    "turntable-justin" = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-warpstream" {
  source = "./modules/repository"

  name                     = "warpstream"
  description              = "Nuon configuration for Warpstream."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    "richardartoul" = "push"
    "ryanworl"      = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-warpstream-installer" {
  source = "./modules/repository"

  name                     = "warpstream-installer"
  description              = "Installer for Warpstream."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {
    "richardartoul" = "push"
    "ryanworl"      = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-okteto" {
  source = "./modules/repository"

  name                     = "okteto"
  description              = "Nuon configuration for Okteto."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
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

  name                     = "meroxa"
  description              = "Nuon configuration + demo for Meroxa."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    simonl2002 = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-weaviate-installer" {
  source = "./modules/repository"

  name                     = "weaviate-installer"
  description              = "Nuon installer Weaviate."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    aduis = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-weaviate" {
  source = "./modules/repository"

  name                     = "weaviate"
  description              = "Nuon configuration + demo for Weaviate."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
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

  name                     = "flipt"
  description              = "Nuon configuration + demo for Flipt."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    georgemac  = "push"
    markphelps = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-honeyhive" {
  source = "./modules/repository"

  name                     = "honeyhive"
  description              = "Nuon configuration + demo for Honeyhive."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  enable_branch_protection = false
  is_private               = true

  collaborators = {
    "michael-hhai" = "push"
    "codehruv"     = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-electric-sql" {
  source = "./modules/repository"

  name                     = "electric-sql"
  description              = "Nuon configuration + demo for Electric SQL."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    thruflo   = "push"
    samwillis = "push"
    alco      = "push"
    balegas   = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-clickhouse" {
  source = "./modules/repository"

  name                     = "clickhouse"
  description              = "Nuon configuration for clickhouse."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-commonfate" {
  source = "./modules/repository"

  name                     = "common-fate"
  description              = "Nuon configuration for common-fate."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    chrnorm           = "push"
    shwethaumashanker = "push"
    JoshuaWilkes      = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-streamkap" {
  source = "./modules/repository"

  name                     = "streamkap"
  description              = "Nuon configuration for streamkap."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    thomasr888      = "push"
    quang-streamkap = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-streamkap-installer" {
  source = "./modules/repository"

  name                     = "streamkap-installer"
  description              = "Custom installer for streamkap."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {
    thomasr888      = "push"
    quang-streamkap = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-100xdev" {
  source = "./modules/repository"

  name                     = "100xdev"
  description              = "Nuon configuration for 100xdev."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    anandsainath     = "push"
    shaumik100x      = "push"
    sraibagiwith100x = "push"
    gyx119           = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-100xdev-installer" {
  source = "./modules/repository"

  name                     = "100xdev-installer"
  description              = "Custom installer for 100xdev."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {
    anandsainath     = "push"
    shaumik100x      = "push"
    sraibagiwith100x = "push"
    gyx119           = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-run-llm" {
  source = "./modules/repository"

  name                     = "run-llm"
  description              = "Nuon configuration for Run LLM."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {}

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-stardog" {
  source = "./modules/repository"

  name                     = "stardog"
  description              = "Nuon configuration for stardog."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    "paulplace" = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-stardog-installer" {
  source = "./modules/repository"

  name                     = "stardog-installer"
  description              = "Custom installer for stardog."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {
    "paulplace" = "push"
  }

  providers = {
    github = github.nuonco-shared
  }
}

module "shared-berri-ai" {
  source = "./modules/repository"

  name                     = "berri-ai"
  description              = "Nuon configuration for Berri AI."
  required_checks          = []
  owning_team_id           = github_team.nuonco-shared.id
  is_private               = true
  enable_branch_protection = false

  collaborators = {
    ishaan-jaff    = "push"
    krrishdholakia = "push"

  }

  providers = {
    github = github.nuonco-shared
  }
}
