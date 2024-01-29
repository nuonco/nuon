module "services" {
  source = "terraform-aws-modules/ecs/aws//modules/service"

  name        = "http_bin"
  cluster_arn = var.cluster_arn

  cpu    = 1024
  memory = 4096

  container_definitions = {

    fluent-bit = {
      cpu       = 512
      memory    = 1024
      essential = true
      image     = "906394416424.dkr.ecr.us-west-2.amazonaws.com/aws-for-fluent-bit:stable"
      firelens_configuration = {
        type = "fluentbit"
      }
      memory_reservation = 50
    }

    ecs-sample = {
      cpu       = 512
      memory    = 1024
      essential = true
      image     = "public.ecr.aws/aws-containers/ecsdemo-frontend:776fd50"
      port_mappings = [
        {
          name          = "ecs-sample"
          containerPort = 80
          protocol      = "tcp"
        }
      ]

      # Example image used requires access to write to root filesystem
      readonly_root_filesystem = false

      dependencies = [{
        containerName = "fluent-bit"
        condition     = "START"
      }]

      # enable_cloudwatch_logging = false
      # log_configuration = {
      #   logDriver = "awsfirelens"
      #   options = {
      #     Name                    = "firehose"
      #     region                  = "eu-west-1"
      #     delivery_stream         = "my-stream"
      #     log-driver-buffer-limit = "2097152"
      #   }
      # }
      memory_reservation = 100
    }
  }

  service_connect_configuration = {
    namespace = "example"
    service = {
      client_alias = {
        port     = 80
        dns_name = "ecs-sample"
      }
      port_name      = "ecs-sample"
      discovery_name = "ecs-sample"
    }
  }

  load_balancer = {
    service = {
      target_group_arn = module.ingress.target_groups["ex-instance"].arn
      container_name   = "ecs-sample"
      container_port   = 80
    }
  }

  subnet_ids = local.subnet_id_list
  security_group_rules = {
    alb_ingress_3000 = {
      type                     = "ingress"
      from_port                = 80
      to_port                  = 80
      protocol                 = "tcp"
      description              = "Service port"
      source_security_group_id = var.ingress_security_group_id
    }
    egress_all = {
      type        = "egress"
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  tags = {
    Environment = "dev"
    Terraform   = "true"
  }
}
