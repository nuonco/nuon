locals {
  subnet_ids = [
    var.subnet_id_one,
    var.subnet_id_two,
  ]
}

module "db" {
  source  = "terraform-aws-modules/rds/aws"
  version = "6.3.0"

  identifier = var.identifier

  engine               = "postgres"
  engine_version       = "14"
  family               = "postgres14"
  major_engine_version = "14"

  parameters = [
    {
      name  = "rds.logical_replication"
      value = 1
    },
  ]

  create_db_option_group    = false
  create_db_parameter_group = false

  instance_class    = "db.t4g.micro"
  allocated_storage = 5

  // in order to set the password ourselves, this must be set to false
  manage_master_user_password = false

  db_name  = var.db_name
  username = var.username
  password = var.password
  port     = var.port

  iam_database_authentication_enabled = true
  apply_immediately                   = true
  vpc_security_group_ids              = [module.security_group_rds.security_group_id]

  db_subnet_group_name = module.subnet_group.db_subnet_group_id
  subnet_ids           = local.subnet_ids
}
