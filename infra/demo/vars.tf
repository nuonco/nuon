locals {
  role_arn = "arn:aws:iam::949309607565:role/terraform"

  networks = {
    sandbox = {
      cidr             = "10.128.0.0/16"
      public_subnets   = ["10.128.0.0/26", "10.128.0.64/26", "10.128.0.128/26"]
      private_subnets  = ["10.128.128.0/24", "10.128.129.0/24", "10.128.130.0/24"]
      database_subnets = ["10.128.131.0/24", "10.128.132.0/24"]
    }
  }

  rds = {
    engine                  = "postgres"
    engine_version          = "15.8"
    family                  = "postgres15"
    major_engine_version    = "15"
    instance_class          = "db.t4g.micro"
    allocated_storage       = "100"
    db_name                 = "demo"
    port                    = 5432
    username                = "demo"
    multi_az                = false
    backup_retention_period = 1
    skip_final_snapshot     = true
    deletion_protection     = false
    storage_encrypted       = false
  }


  tags = {
    terraform = "true"
  }
}
