module "common-fate" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-ecs"
  prefix = "common-fate"
}

module "warpstream" {
  source = "github.com/nuonco/sandboxes//iam-role"
  sandbox = "aws-ecs"
  prefix = "warpstream"
}
