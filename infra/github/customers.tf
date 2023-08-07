module "customer-meroxa" {
  source = "./modules/repository"

  name        = "customer-meroxa"
  description = "Nuon configuration + demo for Meroxa."
}

module "customer-noteable" {
  source = "./modules/repository"

  name        = "customer-meroxa"
  description = "Nuon configuration + demo for Noteable."
}

module "customer-signoz" {
  source = "./modules/repository"

  name        = "customer-signoz"
  description = "Nuon configuration + demo for Signox."
}
