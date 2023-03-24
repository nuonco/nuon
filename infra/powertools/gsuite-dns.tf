resource "google_dns_record_set" "mx" {
  name         = data.google_dns_managed_zone.root.dns_name
  managed_zone = data.google_dns_managed_zone.root.name
  type         = "MX"
  ttl          = 3600

  rrdatas = [
    "1 ASPMX.L.GOOGLE.COM.",
    "5 ALT1.ASPMX.L.GOOGLE.COM.",
    "5 ALT2.ASPMX.L.GOOGLE.COM.",
    "10 ALT3.ASPMX.L.GOOGLE.COM.",
    "10 ALT4.ASPMX.L.GOOGLE.COM."
  ]
}

resource "google_dns_record_set" "dkim" {
  name         = "google._domainkey.${data.google_dns_managed_zone.root.dns_name}"
  managed_zone = data.google_dns_managed_zone.root.name
  type         = "TXT"
  ttl          = 300

  rrdatas = [
    "\"v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCqlY1zEOxQlpOQJeeNgk+fJaMqAdfu1rM2V9Z08JBOkod2WJuxuty9efnqqmUZwIC97xYivWRvK89D2hVUmYO5yYFgInozICkYLjU0PTUnRObs1/rEA62Kqg0BdauIW6a+RlCmv8+0+v0BA8R9nUozw8HGJnDTCzcAT4XNxoE7FQIDAQAB\""
  ]
}

resource "google_dns_record_set" "spf" {
  name         = data.google_dns_managed_zone.root.dns_name
  managed_zone = data.google_dns_managed_zone.root.name
  type         = "TXT"
  ttl          = 86400
  rrdatas = [
    "\"v=spf1 include:_spf.google.com ~all\"",
    "google-site-verification=ESTIshJxUp9wpeJMMCP0IK4w0xTVKRN9tSLJ9049frE"
  ]
}
