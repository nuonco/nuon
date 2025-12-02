locals {
  name                 = "nuonctl"
  bucket_name          = "nuon-build-manifests"
  account_locks_bucket = "nuon-account-locks"

  github = {
    organization = "powertoolsdev"
    repo         = "mono"
  }

  tags = {
    terraform = "infra-nuonctl"
  }
}
