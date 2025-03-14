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
    "services/e2e" = {
      bucket_prefix   = "e2e"
      ecr             = module.e2e.all
      use_promotions  = true
      push_latest_tag = false
    }

    // charts
    "charts/demo" = {
      bucket_prefix   = "helm-demo"
      ecr             = module.helm_demo.all
      use_promotions  = false
      push_latest_tag = false
    }
    "charts/temporal" = {
      bucket_prefix   = "helm-temporal"
      ecr             = module.helm_temporal.all
      use_promotions  = false
      push_latest_tag = false
    }
    "charts/waypoint" = {
      bucket_prefix   = "helm-waypoint"
      ecr             = module.helm_waypoint.all
      use_promotions  = false
      push_latest_tag = false
    }

    // binaries
    "bins/cli" = {
      bucket_prefix   = "cli"
      ecr             = module.cli.all
      use_promotions  = true
      push_latest_tag = false
    }
    "bins/nuonctl" = {
      bucket_prefix  = "nuonctl"
      ecr            = module.nuonctl.all
      use_promotions = false
      # this is mainly to test the functionality
      push_latest_tag = true
    }
    "bins/runner" = {
      bucket_prefix   = "runner"
      ecr             = module.runner.all
      use_promotions  = true
      push_latest_tag = true
    }

    "bins/stage-runner" = {
      bucket_prefix   = "stage-runner"
      ecr             = module.runner.all
      use_promotions  = false
      push_latest_tag = true
    }

    // sandboxes
    "sandboxes/aws-eks" = {
      bucket_prefix   = "sandbox/aws-eks"
      ecr             = module.sandbox_aws_eks.all
      use_promotions  = true
      push_latest_tag = false
    }
    "sandboxes/empty" = {
      bucket_prefix   = "sandbox/empty"
      ecr             = module.sandbox_empty.all
      use_promotions  = true
      push_latest_tag = false
    }
  }
}
