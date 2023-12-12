module "cluster" {
  source  = "cloudposse/documentdb-cluster/aws"
  version = "0.24.0"

  namespace               = "eg"
  stage                   = "testing"
  name                    = "docdb"
  cluster_size            = 3
  master_username         = "admin1"
  master_password         = "Test123456789"
  instance_class          = "db.r4.large"
  vpc_id                  = "vpc-xxxxxxxx"
  subnet_ids              = ["subnet-xxxxxxxx", "subnet-yyyyyyyy"]
  allowed_security_groups = ["sg-xxxxxxxx"]
  zone_id                 = "Zxxxxxxxx"
}
