terraform {
  required_version = ">= 1.3.7"


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
      version = ">= 0.8.0"
    }
  }
}
