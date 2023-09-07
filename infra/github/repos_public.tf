module "quickstart" {
  source          = "./modules/repository"
  name            = "quickstart"
  description     = "A simple example project to easily get up and running with Nuon."
  required_checks = []
  is_public       = true
}

module "terraform-provider-nuon" {
  source          = "./modules/repository"
  name            = "terraform-provider-nuon"
  description     = "A Terraform provider for managing applications in Nuon."
  required_checks = []
  is_public       = true
}

module "nuon-sdk" {
  source          = "./modules/repository"
  name            = "nuon-sdk"
  description     = "An SDK for interacting with the Nuon platform."
  required_checks = []
  is_public       = true
}
