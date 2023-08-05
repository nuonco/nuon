locals {
  bucket_name = "e2e-${var.install_id}"
}

// setup the bucket so that anyone in the account can assume it. A much better solution would be to create an IAM role
// that can only be assumed by the eks service of choice, but that can come later.
data "aws_iam_policy_document" "bucket_access_policy" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
    ]
    resources = ["arn:aws:s3:::${local.bucket_name}", ]
    principals {
      type        = "AWS"
      identifiers = ["*", ]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "s3:*Object",
    ]
    resources = ["arn:aws:s3:::${local.bucket_name}/*", ]
    principals {
      type        = "AWS"
      identifiers = ["*", ]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
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

  control_object_ownership = true
  object_ownership         = "BucketOwnerEnforced"

  attach_policy = true
  policy        = data.aws_iam_policy_document.bucket_access_policy.json
}
