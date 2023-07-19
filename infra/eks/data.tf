data "twingate_groups" "engineers" {
  name = "engineers"
}

// this is the public dns name, created by `infra-aws`
data "aws_route53_zone" "env_root" {
  name = "${var.account}.nuon.co"
}
