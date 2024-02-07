module "eks_access" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-eks"
  prefix = "e2e-jon"
}
