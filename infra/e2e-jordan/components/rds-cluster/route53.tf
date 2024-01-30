resource "aws_route53_record" "primary" {
  zone_id = data.aws_route53_zone.private.id
  name    = "${var.identifier}-primary.db.${data.aws_route53_zone.private.name}"
  type    = "A"

  alias {
    name                   = module.db.db_instance_address
    zone_id                = module.db.db_instance_hosted_zone_id
    evaluate_target_health = false
  }
}
