variable "install_role_arn" {
  description = "install role arn"
}

variable "east_1_count" {
  default = 5
  type = number
}

variable "east_2_count" {
  default = 5
  type = number
}

variable "west_2_count" {
  default = 5
  type = number
}
