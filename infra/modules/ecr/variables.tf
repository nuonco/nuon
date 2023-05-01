variable "name" {
  type = string
}

variable "tags" {
  type = map(any)
}

variable "region" {
  type        = string
  description = "us-west-2"
}
