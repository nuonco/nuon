variable "aws_region" {
  type    = string
  default = "us-west-2"
}

variable "domain" {
  type        = string
  description = "the internal domain you want to make use of."
}

variable "env" {
  type = string
}

variable "subdomain" {
  type = string
}

variable "service" {
  type = string
}
