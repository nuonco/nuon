module "eks_access" {
  source = "nuonco/install-access/aws"
  sandbox = "aws-eks"
  prefix = "e2e-jon"
}
