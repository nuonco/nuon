locals {
  org_id     = data.aws_organizations_organization.orgs.id
  account_id = local.accounts[local.aws_settings.account_name]
  public_prefixes = [
    "cli/*",
    "runner/*",
    "nuonctl/*",
    "terraform-provider-nuon/*",
    "sandbox/*",
    "cfngen/*",
  ]
  replication_prefixes = [
    "cfngen/*",
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

# Cross-region replication setup
data "aws_iam_policy_document" "replication_assume_role" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "replication" {
  name               = "${local.bucket_name}-replication-role"
  assume_role_policy = data.aws_iam_policy_document.replication_assume_role.json
}

data "aws_iam_policy_document" "replication_policy" {
  for_each = toset(local.replication_prefixes)
  statement {
    effect = "Allow"
    actions = [
      "s3:GetReplicationConfiguration",
      "s3:ListBucket"
    ]
    resources = [module.bucket.s3_bucket_arn]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObjectVersionForReplication",
      "s3:GetObjectVersionAcl",
      "s3:GetObjectVersionTagging"
    ]
    resources = ["${module.bucket.s3_bucket_arn}/${each.value}"]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:ReplicateObject",
      "s3:ReplicateDelete",
      "s3:ReplicateTags"
    ]
    resources = [for region in data.aws_regions.current.names : "arn:aws:s3:::${local.bucket_name}-${region}/${each.value}"]
  }
}

resource "aws_iam_role_policy" "replication" {
  for_each = data.aws_iam_policy_document.replication_policy
  name     = "${local.bucket_name}-replication-policy"
  role     = aws_iam_role.replication.id
  policy   = each.value.json
}

# Get list of all AWS regions
data "aws_regions" "current" {
  all_regions = true
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required", "opted-in"]
  }
}

# Create destination buckets in each region
module "destination_buckets" {
  for_each = toset(data.aws_regions.current.names)
  source   = "terraform-aws-modules/s3-bucket/aws"
  version  = ">= v3.2.4"

  bucket = "${local.bucket_name}-${each.key}"
  versioning = {
    enabled = true
  }
  # Mirror the same security settings as source bucket
  attach_deny_insecure_transport_policy = true
  attach_require_latest_tls_policy      = true
  attach_public_policy                  = false
  block_public_acls                     = false
  block_public_policy                   = false
  control_object_ownership              = true
  object_ownership                      = "BucketOwnerEnforced"
}

# Add replication configuration to source bucket
resource "aws_s3_bucket_replication_configuration" "replication" {
  depends_on = [module.bucket]

  role   = aws_iam_role.replication.arn
  bucket = module.bucket.s3_bucket_id

  dynamic "rule" {
    for_each = toset(data.aws_regions.current.names)
    content {
      id     = "replicate-to-${rule.value}"
      status = "Enabled"

      destination {
        bucket        = module.destination_buckets[rule.value].s3_bucket_arn
        storage_class = "STANDARD"
      }
    }
  }
}

# Create Multi-Region Access Point
resource "aws_s3control_multi_region_access_point" "mrap" {
  details {
    name = "${local.bucket_name}-mrap"
    region {
      bucket = module.bucket.s3_bucket_id
    }

    dynamic "region" {
      for_each = toset(data.aws_regions.current.names)
      content {
        bucket = module.destination_buckets[region.value].s3_bucket_id
      }
    }
  }
}
