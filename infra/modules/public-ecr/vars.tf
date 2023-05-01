variable "name" {
  type        = string
  description = "Repository name."
}

variable "about" {
  type        = string
  description = "Markdown about section."
}

variable "description" {
  type        = string
  description = "Repository description."
}

variable "region" {
  type        = string
  description = "us-east-1"
}

variable "tags" {
  type        = map(any)
  description = "Tags to add to resources"
}

variable "logo_image_path" {
  type = string
  # NOTE: if you are using a module outside of `infra/<projects>` you will have to change or override this to hit the
  # default logo, or provide your own image.
  default     = "../modules/public-ecr/logo.png"
  description = "path to the logo image"
}
