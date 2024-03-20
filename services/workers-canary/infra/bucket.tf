locals {
  bucket_name = "nuon-canary-${var.env}"
  org_id = data.aws_organizations_organization.orgs.id
}

resource "aws_kms_key" "canary_bucket" {
  provider = aws.canary

  description = "KMS key for ${local.bucket_name}"
  policy      = data.aws_iam_policy_document.canary_bucket_key_policy.json

  deletion_window_in_days = 7
}

resource "aws_kms_alias" "canary_bucket" {
  provider = aws.canary

  name          = "alias/bucket-key-${local.bucket_name}"
  target_key_id = aws_kms_key.canary_bucket.key_id
}

data "aws_iam_policy_document" "canary_bucket_key_policy" {
  provider = aws.canary

  # enable IAM User Permissions
  statement {
    effect    = "Allow"
    actions   = ["kms:*", ]
    resources = ["*", ]

    principals {
      type        = "AWS"
      identifiers = [local.accounts["canary"].id, ]
    }
  }

  # allow all principals in this account that are authorized for s3
  statement {
    effect = "Allow"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
    resources = ["*", ]
    principals {
      type        = "AWS"
      identifiers = ["*", ]
    }
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["s3.us-west-2.amazonaws.com", ]
    }
    condition {
      test     = "StringEquals"
      variable = "kms:CallerAccount"
      values   = [local.accounts["canary"].id, ]
    }
  }

  # allow all principals in the nuon org that are authorized for s3
  statement {
    effect = "Allow"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
    resources = ["*", ]
    principals {
      type        = "AWS"
      identifiers = ["*", ]
    }
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["s3.us-west-2.amazonaws.com", ]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      values   = [data.aws_organizations_organization.orgs.id]
    }
  }
}

data "aws_iam_policy_document" "canary_bucket_policy" {
  provider = aws.canary

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
      values   = [local.org_id]
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
}

module "canary_bucket" {
  providers = {
    aws = aws.canary
  }
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
  policy        = data.aws_iam_policy_document.canary_bucket_policy.json

  server_side_encryption_configuration = {
    rule : [
      {
        apply_server_side_encryption_by_default : {
          kms_master_key_id = aws_kms_key.canary_bucket.arn
          sse_algorithm : "aws:kms",
        },
        bucket_key_enabled : true,
      },
    ],
  }
}
