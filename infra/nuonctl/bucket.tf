data "aws_iam_policy_document" "bucket_policy" {
  statement {
    sid    = "AllowOrgAccountsAccess"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = [for account_name, account_id in local.accounts : "arn:aws:iam::${account_id}:root"]
    }
    actions = [
      "s3:ListBucket",
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
    ]
    resources = [
      "arn:aws:s3:::${local.bucket_name}",
      "arn:aws:s3:::${local.bucket_name}/*",
    ]
  }
}

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

  attach_policy = true
  policy        = data.aws_iam_policy_document.bucket_policy.json
}

module "account_locks_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  bucket = local.account_locks_bucket
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
}
