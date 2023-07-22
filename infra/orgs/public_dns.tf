# NOTE: we import the public dns zone as a resource here, assuming it is in the public account.
# We do this, to prevent a dependency on the dns infra, and consolidate all the "shared" infra for orgs here.
# However, we ultimately want to centralize all DNS in `infra/dns` since it's easy to mess up globally.
data "aws_route53_zone" "public_domain" {
  provider = aws.public
  name     = local.vars.public_domain
}
