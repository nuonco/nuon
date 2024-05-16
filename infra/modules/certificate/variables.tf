variable "use_root_domain" {
  type    = bool
  default = false
}

variable "aws_region" {
  type    = string
  default = "us-west-2"
}

variable "root_domain" {
  type    = string
  default = "nuon.co"
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
