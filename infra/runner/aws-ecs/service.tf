locals {
  # Saw these in the service.yml for the runner.
  # Not sure if they're required.
  git_ref                  = "local"
  settings_refresh_timeout = "1m"
}

module "service" {
  source = "terraform-aws-modules/ecs/aws//modules/service"

  name          = var.runner_id
  cluster_arn   = var.cluster_arn
  desired_count = 1
  cpu           = 1024
  memory        = 4096

  container_definitions = {
    "${var.runner_id}" = {
      image                    = "${var.image_url}:${var.image_tag}"
      cpu                      = 512
      memory                   = 1024
      essential                = true
      memory_reservation       = 100
      readonly_root_filesystem = true

      environment = [
        {
          name  = "RUNNER_API_URL"
          value = var.api_url
        },
        {
          name  = "RUNNER_API_TOKEN"
          value = var.api_token
        },
        {
          name  = "RUNNER_ID"
          value = var.runner_id
        },
        {
          name  = "GIT_REF"
          value = local.git_ref
        },
        {
          name  = "SETTINGS_REFRESH_TIMEOUT"
          value = local.settings_refresh_timeout
        },
      ]
    }
  }

  subnet_ids = local.private_subnet_ids
  security_group_rules = {
    egress_all = {
      type        = "egress"
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  tags = {
    Terraform   = "true"
    ServiceName = var.runner_id
  }
}
