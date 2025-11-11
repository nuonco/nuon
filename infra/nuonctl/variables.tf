locals {
  name        = "nuonctl"
  bucket_name = "nuon-build-manifests"

  github = {
    organization = "powertoolsdev"
    repo         = "mono"
  }

  tags = {
    terraform = "infra-nuonctl"
  }
}
