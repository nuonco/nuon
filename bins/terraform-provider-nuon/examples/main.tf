terraform {
  required_providers {
    nuon = {
      source = "nuonco/nuon"
    }
  }
}

provider "nuon" {
  # Defaults to https://api.nuon.co
  # api_url = "https://api.nuon.co"
}

resource "nuon_stack" "example" {
  phone_home_id = var.phone_home_id

  aws {
    # region / account_id are resolved from the control plane when omitted.
    state_bucket = "my-nuon-tf-state"
    # state_key defaults to nuon/<install_id>/terraform.tfstate
  }

  inputs = {
    domain = "example.com"
  }

  secrets = {
    db_password = var.db_password
  }
}

variable "phone_home_id" {
  type = string
}

variable "db_password" {
  type      = string
  sensitive = true
}

output "runner_role_arn" {
  value = nuon_stack.example.outputs["runner_iam_role_arn"]
}
