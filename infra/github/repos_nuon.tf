module "quickstart-nuon" {
  source           = "./modules/repository"
  name             = "quickstart"
  description      = "Create a fully managed version of your app that runs in your customer’s cloud account."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

# This had to start as a fork of nuonco/quickstart, so we could test
# using the quickstart in the same way a vendor would.
# Leaving this here as a record of how this repo was created.
import {
  to = module.quickstart_test_nuon.github_repository.main
  id = "quickstart-test"
}

module "quickstart_test_nuon" {
  source           = "./modules/repository"
  name             = "quickstart-test"
  description      = "Repo for testing the quickstart"
  required_checks  = []
  is_public        = true
  is_fork          = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuonco" {
  source           = "./modules/repository"
  name             = ".github"
  description      = "Run your app in your customer's cloud with nuon.co"
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "terraform-provider-nuon" {
  source           = "./modules/repository"
  name             = "terraform-provider-nuon"
  description      = "A Terraform provider for managing applications in Nuon."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-go" {
  source           = "./modules/repository"
  name             = "nuon-go"
  description      = "An SDK for interacting with the Nuon platform."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-runner-go" {
  source           = "./modules/repository"
  name             = "nuon-runner-go"
  description      = "The SDK that powers the Nuon runner."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-elixir" {
  source      = "./modules/repository"
  name        = "nuon-elixir"
  description = "An SDK for integrating with Nuon from Elixir."
  required_checks = [
    "check-pr / Run PR checks",
    "check-pr / Update PR status",
  ]
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-python" {
  source      = "./modules/repository"
  name        = "nuon-python"
  description = "An SDK for integrating with Nuon from Python."
  required_checks = [
    "check-pr / Run PR checks",
    "check-pr / Update PR status",
  ]
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-actions-build" {
  source      = "./modules/repository"
  name        = "actions-build"
  description = "Action for building a Nuon component."
  required_checks = [
    "check-pr / Run PR checks",
    "check-pr / Update PR status",
  ]
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-actions-release" {
  source      = "./modules/repository"
  name        = "actions-release"
  description = "Action for releasing a Nuon build."
  required_checks = [
    "check-pr / Run PR checks",
    "check-pr / Update PR status",
  ]
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-guides" {
  source      = "./modules/repository"
  name        = "guides"
  description = "Project code for guides."
  required_checks = [
    "check-pr / Run PR checks",
    "check-pr / Update PR status",
  ]
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-homebrew-tap" {
  source      = "./modules/repository"
  name        = "homebrew-tap"
  description = "Homebrew tap for the Nuon CLI."
  required_checks = [
    "check-pr / Run PR checks",
    "check-pr / Update PR status",
    "test-bot (ubuntu-22.04)",
    "test-bot (macos-13)",
  ]
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-terraform-aws-ecr-access" {
  source           = "./modules/repository"
  name             = "terraform-aws-ecr-access"
  description      = "Terraform module for granting access for Nuon container image components."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-terraform-aws-access-delegation" {
  source           = "./modules/repository"
  name             = "terraform-aws-install-access-delegation"
  description      = "Set up an IAM role that allows you to setup a delegation IAM role using Nuon."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-terraform-aws-install-access" {
  source           = "./modules/repository"
  name             = "terraform-aws-install-access"
  description      = "Terraform module for granting access to Nuon to provision an install."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-vpn-configuration-examples" {
  source           = "./modules/repository"
  name             = "vpn-configuration-examples"
  description      = "Examples for configuring VPNs, for connecting to BYOC deployed applications."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-terraform-installer-ui" {
  source           = "./modules/repository"
  name             = "installer"
  description      = "Installer UI"
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-demo" {
  source           = "./modules/repository"
  name             = "demo"
  description      = "Demo app built with Nuon."
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-demo-installer" {
  source = "./modules/repository"

  name                     = "demo-installer"
  description              = "Demo installer fork."
  required_checks          = []
  owning_team_id           = github_team.nuon.id
  is_public                = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {
  }

  providers = {
    github = github.nuon
  }
}

module "kitchen-sink-installer" {
  source = "./modules/repository"

  name                     = "kitchen-sink-installer"
  description              = "Kitchen Sink installer fork."
  required_checks          = []
  owning_team_id           = github_team.nuon.id
  is_public                = true
  enable_branch_protection = false
  is_fork                  = true

  collaborators = {
  }

  providers = {
    github = github.nuon
  }
}

module "nuon-examples" {
  source = "./modules/repository"

  name                     = "nuon-examples"
  description              = "Example Application Configurations"
  required_checks          = []
  owning_team_id           = github_team.nuon.id
  is_public                = true
  enable_branch_protection = false

  collaborators = {
  }

  providers = {
    github = github.nuon
  }
}

module "nuon-terraform-aws-vpc" {
  source           = "./modules/repository"
  name             = "terraform-aws-vpc"
  description      = "Terraform module for creating a VPC to install BYOPC installs into."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-components" {
  source           = "./modules/repository"
  name             = "components"
  description      = "Library of common components for use with nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-actions" {
  source           = "./modules/repository"
  name             = "actions"
  description      = "Library of common actions for use with nuon apps."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-eks-permissions" {
  source           = "./modules/repository"
  name             = "eks-permissions"
  description      = "Basic permissions for managing an EKS install."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "nuon-policies" {
  source           = "./modules/repository"
  name             = "policies"
  description      = "Default policies for sandboxes and runner-jobs."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "content" {
  source           = "./modules/repository"
  name             = "content"
  description      = "Scratch pad for writing content, blogs and more."
  required_checks  = []
  is_public        = false
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "aws-cloudformation-templates" {
  source           = "./modules/repository"
  name             = "aws-cloudformation-templates"
  description      = "Public templates that can be used within Nuon CloudFormation stacks."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "runner" {
  source           = "./modules/repository"
  name             = "runner"
  description      = "Public components and supporting tools for the Nuon runner."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}

module "byoc" {
  source           = "./modules/repository"
  name             = "byoc"
  description      = "Nuon, but make it BYOC."
  required_checks  = []
  is_public        = true
  owning_team_id   = github_team.nuon.id
  owning_team_name = "nuonco/${github_team.nuon.name}"

  providers = {
    github = github.nuon
  }
}
