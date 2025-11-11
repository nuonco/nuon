locals {
  name                   = "artifacts"
  bucket_name            = "nuon-artifacts"
  terraform_organization = "nuonco"

  github = {
    organization = "powertoolsdev"
    repo         = "mono"
  }

  tags = {
    terraform = "infra-artifacts"
  }

  replicated_tags = {
    terraform = "infra-artifacts-replicated"
  }
}
