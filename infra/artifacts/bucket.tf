locals {
  org_id     = data.aws_organizations_organization.orgs.id
  account_id = local.accounts[local.aws_settings.account_name]
  public_prefixes = [
    "aws-cloudformation-templates/*",
    "cli/*",
    "runner/*",
    "nuonctl/*",
    "terraform-provider-nuon/*",
    "sandbox/*",
    "stacks/*",
  ]

}

# give all accounts in our org access to this bucket
data "aws_iam_policy_document" "s3_bucket_policy" {
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
      variable = "aws:PrincipalOrgID"
      values   = [local.org_id, ]
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
      variable = "aws:PrincipalOrgID"
      values   = [local.org_id]
    }
  }

  // allow a few select public paths in the artifacts bucket
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
    ]
    resources = formatlist("arn:aws:s3:::${local.bucket_name}/%s", local.public_prefixes)
    principals {
      type        = "*"
      identifiers = ["*", ]
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
    ]
    resources = [
      "arn:aws:s3:::${local.bucket_name}",
    ]
    principals {
      type        = "*"
      identifiers = ["*", ]
    }
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = local.public_prefixes
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
  block_public_acls    = false
  block_public_policy  = false

  control_object_ownership = true
  object_ownership         = "BucketOwnerEnforced"

  attach_policy = true
  policy        = data.aws_iam_policy_document.s3_bucket_policy.json
}
