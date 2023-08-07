module "customer-meroxa" {
  source = "./modules/repository"

  name        = "customer-meroxa"
  description = "Nuon configuration + demo for Meroxa."
  required_checks = []
}

module "customer-noteable" {
  source = "./modules/repository"

  name        = "customer-noteable"
  description = "Nuon configuration + demo for Noteable."
  required_checks = []
}

module "customer-signoz" {
  source = "./modules/repository"

  name        = "customer-signoz"
  description = "Nuon configuration + demo for Signox."
  required_checks = []
}
