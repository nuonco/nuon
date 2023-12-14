module "cluster" {
  source  = "cloudposse/documentdb-cluster/aws"
  version = "0.24.0"

  namespace           = var.namespace
  stage               = var.stage
  name                = var.name
  cluster_size        = var.cluster_size
  master_username     = var.master_username
  master_password     = var.master_password
  vpc_id              = var.vpc_id
  allowed_cidr_blocks = [data.aws_vpc.vpc.cidr_block_associations[0].cidr_block]
  subnet_ids = [
    var.subnet_one,
    var.subnet_two,
  ]
  zone_id = var.zone_id
}
