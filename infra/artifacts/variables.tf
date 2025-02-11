locals {
  name        = "artifacts"
  bucket_name = "nuon-artifacts"

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
