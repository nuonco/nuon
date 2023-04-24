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
    nuonctl = {
      ecr_repository_url = ""
      ecr_registry_id    = ""
      ecr_is_public      = false
    }

    waypoint = {
      ecr_repository_arn = ""
    }

    temporal = {
      ecr_repository_arn = ""
    }

    sandbox_aws_eks = {
      ecr_repository_url = ""
      ecr_registry_id    = ""
      ecr_is_public      = false
    }

    sandbox_empty = {
      ecr_repository_url = ""
      ecr_registry_id    = ""
      ecr_is_public      = false
    }
  }
}
