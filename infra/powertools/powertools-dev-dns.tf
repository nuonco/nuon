data "google_dns_managed_zone" "root" {
  name = "powertools-internal-powertools-dev"
}

resource "google_dns_record_set" "root_testing_delegation" {
  name         = "testing.${var.root_domain}."
  type         = "NS"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas = [
    "ns-cloud-b1.googledomains.com.",
    "ns-cloud-b2.googledomains.com.",
    "ns-cloud-b3.googledomains.com.",
    "ns-cloud-b4.googledomains.com.",
  ]
}

// naked redirect which points to our nakedssl.com account
resource "google_dns_record_set" "naked_ssl_redirect" {
  name         = "${var.root_domain}."
  type         = "A"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas      = ["54.86.103.22"]
}

// sendgrid redirects
resource "google_dns_record_set" "sendgrid_redirect_cname_1" {
  name         = "url7288.${var.root_domain}."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas      = ["sendgrid.net."]
}

resource "google_dns_record_set" "sendgrid_redirect_cname_2" {
  name         = "16972801.${var.root_domain}."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas      = ["sendgrid.net."]
}

// sendgrid email configuration
resource "google_dns_record_set" "sendgrid_cname_1" {
  name         = "em2225.${var.root_domain}."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas      = ["u16972801.wl052.sendgrid.net."]
}

resource "google_dns_record_set" "sendgrid_cname_2" {
  name         = "s1._domainkey.${var.root_domain}."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas      = ["s1.domainkey.u16972801.wl052.sendgrid.net."]
}

resource "google_dns_record_set" "sendgrid_cname_3" {
  name         = "s2._domainkey.${var.root_domain}."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas      = ["s2.domainkey.u16972801.wl052.sendgrid.net."]
}

resource "google_dns_record_set" "sendgrid_cname_4" {
  name         = "url631.${var.root_domain}."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas      = ["sendgrid.net."]
}

// NOTE: we disable this _after_ verifying our domain, so we can use the redirecting service
// mailing.powertools.dev redirects
// resource "google_dns_record_set" "sendgrid_cname_8" {
//  name = "mailing.${var.root_domain}."
//  type = "CNAME"
//  ttl = "3600"
//  managed_zone = data.google_dns_managed_zone.root.name
//  rrdatas = ["sendgrid.net."]
//}

resource "google_dns_record_set" "sendgrid_cname_9" {
  name         = "16972801.${var.root_domain}."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.root.name
  rrdatas      = ["sendgrid.net."]
}
