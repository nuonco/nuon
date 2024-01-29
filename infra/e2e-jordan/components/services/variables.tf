locals {
  subnet_id_list = split(",", var.subnet_ids)
}

variable "cluster_arn" {
  type = string
}

variable "subnet_ids" {
  type = string
}

variable "region" {
  type = string
}

variable "aws_account_id" {
  type = string
}

variable "ingress_security_group_id" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "target_group_arn" {
  type = string
}
