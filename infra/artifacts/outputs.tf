output "gh_role_arn" {
  value       = module.github_actions.iam_role_arn
  description = "github role"
}

output "region" {
  value = local.aws_settings.region
}

// NOTE(jm): all artifacts share the same bucket, but have different ECR repos for artifacts.
output "bucket" {
  value = {
    name   = module.bucket.s3_bucket_id
    arn    = module.bucket.s3_bucket_arn
    region = module.bucket.s3_bucket_region
  }
}

output "artifacts" {
  value = {
    // e2e
    "e2e" = {
      bucket_prefix   = "e2e"
      ecr             = module.e2e.all
    }

    // binaries
    "cli" = {
      bucket_prefix   = "cli"
      ecr             = module.cli.all
    }

    "nuonctl" = {
      bucket_prefix  = "nuonctl"
      ecr            = module.nuonctl.all
    }

    "runner" = {
      bucket_prefix   = "runner"
      ecr             = module.runner.all
    }

    "website" = {
      bucket_prefix   = "website"
      ecr             = module.website.all
    }

    "docs" = {
      bucket_prefix   = "docs"
      ecr             = module.website.all
    }
  }
}
