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
      bucket_prefix  = "e2e"
      ecr            = module.e2e.all
      use_promotions = true
    }

    // charts
    "charts/demo" = {
      bucket_prefix  = "helm-demo"
      ecr            = module.helm_demo.all
      use_promotions = false
    }
    "charts/temporal" = {
      bucket_prefix  = "helm-temporal"
      ecr            = module.helm_temporal.all
      use_promotions = false
    }
    "charts/waypoint" = {
      bucket_prefix  = "helm-waypoint"
      ecr            = module.helm_waypoint.all
      use_promotions = false
    }

    // binaries
    "bins/cli" = {
      bucket_prefix  = "cli"
      ecr            = module.cli.all
      use_promotions = true
    }
    "bins/nuonctl" = {
      bucket_prefix  = "nuonctl"
      ecr            = module.nuonctl.all
      use_promotions = false
    }
    "bins/runner" = {
      bucket_prefix  = "runner"
      ecr            = module.runner.all
      use_promotions = true
    }
    "bins/waypoint-plugin-exp" = {
      bucket_prefix  = "waypoint-plugin-exp"
      ecr            = module.waypoint_plugin_exp.all
      use_promotions = false
    }
    "bins/waypoint-plugin-helm" = {
      bucket_prefix  = "waypoint-plugin-helm"
      ecr            = module.waypoint_plugin_helm.all
      use_promotions = false
    }
    "bins/waypoint-plugin-noop" = {
      bucket_prefix  = "waypoint-plugin-noop"
      ecr            = module.waypoint_plugin_noop.all
      use_promotions = false
    }
    "bins/waypoint-plugin-oci" = {
      bucket_prefix  = "waypoint-plugin-oci"
      ecr            = module.waypoint_plugin_oci.all
      use_promotions = false
    }
    "bins/waypoint-plugin-oci-sync" = {
      bucket_prefix  = "waypoint-plugin-oci-sync"
      ecr            = module.waypoint_plugin_oci_sync.all
      use_promotions = false
    }
    "bins/waypoint-plugin-terraform" = {
      bucket_prefix  = "waypoint-plugin-terraform"
      ecr            = module.waypoint_plugin_terraform.all
      use_promotions = false
    }
    "bins/waypoint-plugin-job" = {
      bucket_prefix  = "waypoint-plugin-job"
      ecr            = module.waypoint_plugin_job.all
      use_promotions = false
    }

    // sandboxes
    "sandboxes/aws-eks" = {
      bucket_prefix  = "sandbox/aws-eks"
      ecr            = module.sandbox_aws_eks.all
      use_promotions = true
    }
    "sandboxes/empty" = {
      bucket_prefix  = "sandbox/empty"
      ecr            = module.sandbox_empty.all
      use_promotions = true
    }
  }
}
