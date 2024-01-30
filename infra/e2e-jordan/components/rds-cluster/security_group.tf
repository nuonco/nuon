module "security_group_rds" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 5.0"

  name        = var.identifier
  description = "RDS security group for ${var.identifier}"
  vpc_id      = data.aws_vpc.vpc.id

  # ingress
  ingress_with_cidr_blocks = [
    {
      from_port   = var.port
      to_port     = var.port
      protocol    = "tcp"
      description = "RDS access from within VPC"
      cidr_blocks = data.aws_vpc.vpc.cidr_block_associations[0].cidr_block
    },
  ]
}
