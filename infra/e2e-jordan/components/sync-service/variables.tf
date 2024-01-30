locals {
  subnet_id_list = split(",", trim(var.subnet_ids, "[]"))
}

# Service config

variable "database_url" {
  type = string
}

variable "auth_mode" {
  type = string
}

variable "pg_proxy_password" {
  type = string
}


# Hosting config

variable "vpc_id" {
  type = string
}

variable "cluster_arn" {
  type = string
}


# Networking config

variable "subnet_ids" {
  type = string
}

variable "domain_name" {
  type = string
}

variable "zone_id" {
  type = string
}
