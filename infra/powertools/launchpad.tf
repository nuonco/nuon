// launchpad nameserver configuration
// NOTE: we do not import resources from our launchpad AWS account to prevent a dependency between that infrastructure
// and this plan.
resource "google_dns_record_set" "launchpad_prod_delegation" {
  name         = "internal.${var.root_domain}."
  type         = "NS"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas = [
    "ns-1051.awsdns-03.org.",
    "ns-90.awsdns-11.com.",
    "ns-516.awsdns-00.net.",
    "ns-1610.awsdns-09.co.uk.",
  ]
}

resource "google_dns_record_set" "launchpad_stage_delegation" {
  name         = "stage.${var.root_domain}."
  type         = "NS"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas = [
    "ns-270.awsdns-33.com.",
    "ns-692.awsdns-22.net.",
    "ns-1235.awsdns-26.org.",
    "ns-1642.awsdns-13.co.uk.",
  ]
}
