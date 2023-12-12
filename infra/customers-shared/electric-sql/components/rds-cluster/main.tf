module "db" {
  source  = "terraform-aws-modules/rds/aws"
  version = "6.3.0"

  identifier = var.identifier

  engine            = var.engine
  engine_version    = var.engine_version
  instance_class    = var.instance_class
  allocated_storage = 5

  db_name  = var.db_name
  username = var.username
  password = var.password
  port     = var.port

  iam_database_authentication_enabled = var.iam_database_authentication_enabled
  vpc_security_group_ids              = var.vpc_security_group_ids

  create_db_subnet_group = true
  subnet_ids             = var.subnet_ids

  tags = var.tags
}
