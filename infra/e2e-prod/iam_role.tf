module "install_access" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-eks"
  prefix = "e2e-stage"
}
