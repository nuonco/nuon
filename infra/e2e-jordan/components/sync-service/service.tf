module "service" {
  source = "terraform-aws-modules/ecs/aws//modules/service"

  name        = "http_bin"
  cluster_arn = var.cluster_arn

  cpu    = 1024
  memory = 4096

  container_definitions = {
    electric-sql = {
      cpu       = 512
      memory    = 1024
      essential = true
      image     = "docker.io/electricsql/electric"
      port_mappings = [
        {
          name          = "satellite-http"
          containerPort = 5133
          protocol      = "tcp"
        },
        {
          name          = "logical-publisher-tcp"
          containerPort = 5433
          protocol      = "tcp"
        },
        {
          name          = "pg-proxy-tcp"
          containerPort = 65432
          protocol      = "tcp"
        }
      ]
      memory_reservation = 100
      environment = [
        {
          name  = "DATABASE_URL"
          value = var.database_url
        },
        {
          name  = "AUTH_MODE"
          value = var.auth_mode
        },
        {
          name  = "LOGICAL_PUBLISHER_HOST"
          value = var.domain_name
        },
        {
          name  = "PG_PROXY_PASSWORD"
          value = var.pg_proxy_password
        },
      ]
    }
  }

  service_connect_configuration = {
    namespace = "electric-sql"
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
      container_name   = "electric-sql"
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
      source_security_group_id = module.ingress.security_group_id
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
