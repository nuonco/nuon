module "bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  bucket = local.bucket_name
  versioning = {
    enabled = true
  }

  attach_deny_insecure_transport_policy = true
  attach_require_latest_tls_policy      = true

  attach_public_policy = false
  block_public_acls    = true
  block_public_policy  = true

  control_object_ownership = true
  object_ownership         = "BucketOwnerEnforced"

  # Access is controlled via IAM role policies in:
  # - infra/modules/service/github.tf (for service builds)
  # - infra/artifacts/github_actions.tf (for artifact builds)
  # No need for org-wide bucket policy
  attach_policy = false
}
