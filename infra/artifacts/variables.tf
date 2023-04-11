locals {
  name                   = "artifacts"
  terraform_organization = "launchpaddev"
  bucket_name = "nuon-artifacts"

  github = {
    organization = "powertoolsdev"
    repo = "mono"
  }

  tags = {
    terraform   = "infra-artifacts"
  }
}
