variable "sandbox_org_id" {
  description = "sandbox org id"
}

variable "disable_installs" {
  description = "disable all installs, even if declared in real.yml"
  type = bool
  default = false
}

variable "org_id" {
  description = "org id"
}
