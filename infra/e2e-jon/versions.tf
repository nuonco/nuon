terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      name = "e2e-jon"
    }
  }

  # NOTE: uncomment this to run locally using `nuonctl scripts exec install-terraform-provider`
  #required_providers {
    #nuon = {
      #source  = "terraform.local/local/nuon"
      #version = "0.0.1"
    #}
  #}

  required_providers {
    nuon = {
      source  = "nuonco/nuon"
      version = ">= 0.9.2"
    }
  }
}
