variable "install_id" {
  type        = string
  description = "install_id"
}

variable "region" {
  type        = string
  description = "region"
}

variable "public_domain" {
  type        = string
  description = "the root domain for the install. Usually {{.nuon.install.public_domain}}"
}

variable "public_domain_zone_id" {
  type        = string
  description = "the zone id for the public domain for the install"
}
