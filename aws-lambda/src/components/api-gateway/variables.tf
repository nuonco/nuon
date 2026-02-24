locals {
  fqdn = "${var.name}.${var.domain_name}"
}

variable "region" {
  type = string
}

variable "install_id" {
  type = string
}

variable "name" {
  type = string

  description = "used to name resources and as the subdomain"
}

variable "lambda_function_arn" {
  type = string
}

variable "domain_name" {
  type        = string
  description = "the root domain name"
}


variable "domain_name_certificate_arn" {
  type = string
}

variable "zone_id" {
  type = string
}
