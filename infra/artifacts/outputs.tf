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
    demo = {
      ecr = {
        repository_url = module.helm_demo.repository_url
        registry_id    = module.helm_demo.registry_id
        repository_arn = module.helm_demo.repository_arn
        is_public      = module.helm_demo.is_public
        region         = module.helm_demo.region
        registry_url   = module.helm_demo.registry_url
      }
    }

    nuonctl = {
      ecr = {
        repository_url = module.nuonctl.repository_url
        registry_id    = module.nuonctl.registry_id
        registry_url   = module.nuonctl.registry_url
        repository_arn = module.nuonctl.repository_arn
        is_public      = module.nuonctl.is_public
        region         = module.nuonctl.region
      }
    }

    waypoint-plugin-exp = {
      ecr = {
        repository_url = module.waypoint_plugin_exp.repository_url
        registry_id    = module.waypoint_plugin_exp.registry_id
        registry_url   = module.waypoint_plugin_exp.registry_url
        repository_arn = module.waypoint_plugin_exp.repository_arn
        is_public      = module.waypoint_plugin_exp.is_public
        region         = module.waypoint_plugin_exp.region
      }
    }

    waypoint-plugin-terraform = {
      ecr = {
        repository_url = module.waypoint_plugin_terraform.repository_url
        registry_id    = module.waypoint_plugin_terraform.registry_id
        repository_arn = module.waypoint_plugin_terraform.repository_arn
        is_public      = module.waypoint_plugin_terraform.is_public
        region         = module.waypoint_plugin_terraform.region
        registry_url   = module.waypoint_plugin_terraform.registry_url
      }
    }
  }
}
