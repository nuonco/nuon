locals {
  # Common bucket configuration
  bucket_config = {
    versioning = {
      enabled = true
    }
    attach_deny_insecure_transport_policy = true
    attach_require_latest_tls_policy      = true
    attach_public_policy                  = false
    block_public_acls                     = false
    block_public_policy                   = false
    control_object_ownership              = true
    object_ownership                      = "BucketOwnerEnforced"
    attach_policy                         = true
    policy                                = data.aws_iam_policy_document.s3_bucket_policy.json
  }
}

# This file creates nuon-artifacts buckets across 16 regions, encapsulated
# within a multi-region access point, where the subpath stacks/* is
# replicated from the main nuon-artifacts bucket.

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
    resources = ["${module.bucket.s3_bucket_arn}/*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:ReplicateObject",
      "s3:ReplicateDelete",
      "s3:ReplicateTags"
    ]
    resources = [for region in local.replication_regions : "arn:aws:s3:::${local.bucket_name}-${region}/stacks/*"]
  }
}

resource "aws_iam_policy" "replication" {
  name   = "${local.bucket_name}-replication-policy"
  policy = data.aws_iam_policy_document.replication_policy.json
}

resource "aws_iam_role_policy_attachment" "replication" {
  role       = aws_iam_role.replication.name
  policy_arn = aws_iam_policy.replication.arn
}

# Add replication configuration to source bucket
resource "aws_s3_bucket_replication_configuration" "replication" {
  depends_on = [module.bucket]

  role   = aws_iam_role.replication.arn
  bucket = module.bucket.s3_bucket_id

  dynamic "rule" {
    for_each = local.replication_regions
    content {
      id     = "replicate-to-${rule.value}"
      status = "Enabled"
      filter {
        prefix = "stacks/"
      }
      existing_object_replication {
        status = "Enabled"
      }
      priority = tonumber(rule.key)

      destination {
        bucket        = "arn:aws:s3:::${local.bucket_name}-${rule.value}"
        storage_class = "STANDARD"
      }
    }
  }
}

# Create Multi-Region Access Point
resource "aws_s3control_multi_region_access_point" "mrap" {
  details {
    name = "${local.bucket_name}-mrap"

    public_access_block {
      block_public_acls       = false
      block_public_policy     = false
      ignore_public_acls      = false
      restrict_public_buckets = false
    }

    region {
      bucket = module.bucket.s3_bucket_id
    }

    dynamic "region" {
      for_each = toset(local.replication_regions)
      content {
        bucket = "${local.bucket_name}-${region.value}"
      }
    }
  }
}

module "bucket_ap_south_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-ap-south-1
  }

  bucket = "${local.bucket_name}-ap-south-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_eu_north_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-eu-north-1
  }

  bucket = "${local.bucket_name}-eu-north-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_eu_west_3" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-eu-west-3
  }

  bucket = "${local.bucket_name}-eu-west-3"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_eu_west_2" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-eu-west-2
  }

  bucket = "${local.bucket_name}-eu-west-2"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_eu_west_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-eu-west-1
  }

  bucket = "${local.bucket_name}-eu-west-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_ap_northeast_3" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-ap-northeast-3
  }

  bucket = "${local.bucket_name}-ap-northeast-3"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_ap_northeast_2" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-ap-northeast-2
  }

  bucket = "${local.bucket_name}-ap-northeast-2"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_ap_northeast_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-ap-northeast-1
  }

  bucket = "${local.bucket_name}-ap-northeast-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_ca_central_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-ca-central-1
  }

  bucket = "${local.bucket_name}-ca-central-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_sa_east_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-sa-east-1
  }

  bucket = "${local.bucket_name}-sa-east-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_ap_southeast_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-ap-southeast-1
  }

  bucket = "${local.bucket_name}-ap-southeast-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_ap_southeast_2" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-ap-southeast-2
  }

  bucket = "${local.bucket_name}-ap-southeast-2"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_eu_central_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-eu-central-1
  }

  bucket = "${local.bucket_name}-eu-central-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_us_east_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-us-east-1
  }

  bucket = "${local.bucket_name}-us-east-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_us_east_2" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-us-east-2
  }

  bucket = "${local.bucket_name}-us-east-2"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_us_west_1" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-us-west-1
  }

  bucket = "${local.bucket_name}-us-west-1"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}

module "bucket_us_west_2" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= v3.2.4"

  providers = {
    aws = aws.replicated-us-west-2
  }

  bucket = "${local.bucket_name}-us-west-2"

  versioning                            = local.bucket_config.versioning
  attach_deny_insecure_transport_policy = local.bucket_config.attach_deny_insecure_transport_policy
  attach_require_latest_tls_policy      = local.bucket_config.attach_require_latest_tls_policy
  attach_public_policy                  = local.bucket_config.attach_public_policy
  block_public_acls                     = local.bucket_config.block_public_acls
  block_public_policy                   = local.bucket_config.block_public_policy
  control_object_ownership              = local.bucket_config.control_object_ownership
  object_ownership                      = local.bucket_config.object_ownership
  attach_policy                         = local.bucket_config.attach_policy
}
