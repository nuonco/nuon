module "quickstart" {
  source          = "./modules/repository"
  name            = "quickstart"
  description     = "A simple example project to easily get up and running with Nuon."
  required_checks = []
  is_public = true
}
