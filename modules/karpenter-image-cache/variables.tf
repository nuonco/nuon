variable "images" {
  description = "List of container images to pre-cache on the EBS snapshot"
  type        = list(string)
}

variable "vpc_id" {
  description = "VPC ID where the builder instance will launch"
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID for the builder instance. Must have outbound internet access (NAT gateway) for pulling images from public registries."
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
}

variable "volume_size_gb" {
  description = "Size of the image cache EBS volume in GB. Must be large enough to hold all specified container images (uncompressed)."
  type        = number
  default     = 100
}

variable "volume_type" {
  description = "EBS volume type for the cache volume"
  type        = string
  default     = "gp3"
}

variable "instance_type" {
  description = "EC2 instance type for the builder. Larger instances pull images faster due to higher network bandwidth."
  type        = string
  default     = "m5.xlarge"
}

variable "tags" {
  description = "Additional tags applied to all resources"
  type        = map(string)
  default     = {}
}
