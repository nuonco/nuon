// https://registry.terraform.io/modules/vantage-sh/vantage-integration/aws/latest
module "vantage-integration" {
  source  = "vantage-sh/vantage-integration/aws"
  # provisioned with private acl's and only accessed by Vantage via the provisioned cross account role.
  cur_bucket_name = "nuon-vantage-cur"

  tags = local.tags
}
