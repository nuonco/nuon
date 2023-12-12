variable "namespace" {
  type = string
}

variable "stage" {
  type = string
}

variable "name" {
  type = string
}

variable "cluster_size" {
  type = number
}

variable "master_username" {
  type = number
}

variable "master_password" {
  type = number
}

variable "instance_class" {
  type = number
}

variable "vpc_id" {
  type = number
}

variable "subnet_ids" {
  type = list(string)
}

variable "allowed_security_groups" {
  type = list(string)
}

variable "zone_id" {
  type = number
}
