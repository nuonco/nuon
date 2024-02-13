locals {
  db_password = jsondecode(data.aws_secretsmanager_secret_version.db_instance_password.secret_string).password
}

data "aws_secretsmanager_secret_version" "db_instance_password" {
  secret_id = module.primary.db_instance_master_user_secret_arn
}
