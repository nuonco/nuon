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
    // charts
    "charts/demo" = {
      bucket_prefix = "helm-demo"
      ecr           = module.helm_demo.all
    }
    "charts/temporal" = {
      bucket_prefix = "helm-temporal"
      ecr           = module.helm_temporal.all
    }
    "charts/waypoint" = {
      bucket_prefix = "helm-waypoint"
      ecr           = module.helm_waypoint.all
    }

    // binaries
    "bins/nuonctl" = {
      bucket_prefix = "nuonctl"
      ecr           = module.nuonctl.all
    }
    "bins/waypoint-plugin-exp" = {
      bucket_prefix = "waypoint-plugin-exp"
      ecr           = module.waypoint_plugin_exp.all
    }
    "bins/waypoint-plugin-terraform" = {
      bucket_prefix = "waypoint-plugin-terraform"
      ecr           = module.waypoint_plugin_terraform.all
    }
    "bins/waypoint-plugin-noop" = {
      bucket_prefix = "waypoint-plugin-noop"
      ecr           = module.waypoint_plugin_noop.all
    }
  }
}
