variable "powertools_app_id" {
  type = string
}

variable "powertools_app_installation_id" {
  type = string
}

variable "powertools_app_pem_file" {
  type      = string
  sensitive = true
}
