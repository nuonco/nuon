terraform {
  required_version = ">= 1.3.7"

  #backend "remote" {
    #organization = "nuonco"

    #workspaces {
      #prefix = "customers-shared-"
    #}
  #}

  required_providers {
    nuon = {
      source  = "nuonco/nuon"
      version = ">= 0.3.2"
    }

    utils = {
      source  = "cloudposse/utils"
      version = ">= 0.17.23"
    }
  }
}
