data "google_dns_managed_zone" "secondary" {
  name = "powertools-internal-powertoolsdev-info"
}

// sendgrid email configuration
resource "google_dns_record_set" "secondary_sendgrid_cname_1" {
  name         = "em3445.powertoolsdev.info."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.secondary.name
  rrdatas      = ["u16972801.wl052.sendgrid.net."]
}

resource "google_dns_record_set" "secondary_sendgrid_cname_2" {
  name         = "s1._domainkey.powertoolsdev.info."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.secondary.name
  rrdatas      = ["s1.domainkey.u16972801.wl052.sendgrid.net."]
}

resource "google_dns_record_set" "secondary_sendgrid_cname_3" {
  name         = "s2._domainkey.powertoolsdev.info."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.secondary.name
  rrdatas      = ["s2.domainkey.u16972801.wl052.sendgrid.net."]
}

resource "google_dns_record_set" "secondary_sendgrid_cname_4" {
  name         = "url9524.powertoolsdev.info."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.secondary.name
  rrdatas      = ["sendgrid.net."]
}

resource "google_dns_record_set" "secondary_sendgrid_cname_5" {
  name         = "16972801.powertoolsdev.info."
  type         = "CNAME"
  ttl          = "3600"
  managed_zone = data.google_dns_managed_zone.secondary.name
  rrdatas      = ["sendgrid.net."]
}
