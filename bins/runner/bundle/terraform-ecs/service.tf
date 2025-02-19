module "service" {
  source  = "terraform-aws-modules/ecs/aws//modules/service"
  version = "5.12.0"

  name          = var.runner_id
  cluster_arn   = var.cluster_arn
  desired_count = 1
  cpu           = 1024
  memory        = 4096


  create_tasks_iam_role = false
  tasks_iam_role_arn    = var.runner_iam_role_arn # as opposed to task_exec_iam_role_arn

  container_definitions = {
    (var.runner_id) = {
      image                    = "${var.image_url}:${var.image_tag}"
      cpu                      = 512
      memory                   = 1024
      essential                = true
      memory_reservation       = 100
      readonly_root_filesystem = false
      entryPoint               = "/bin/runner"
      command                  = ["run"]
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
          name  = "RUNNER_VERSION"
          value = var.image_tag
        },
        {
          name  = "HOST_IP"
          value = "0.0.0.0"
        },
        {
          name  = "AWS_REGION"
          value = var.aws_region
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
