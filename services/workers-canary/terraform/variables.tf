variable "aws_eks_iam_role_arn" {
  description = "IAM role for AWS EKS sandbox"
}

variable "aws_ecs_iam_role_arn" {
  description = "IAM role for AWS ECS sandbox"
}

variable "azure_aks_subscription_id" {
  description = "Azure AKS subscription ID"
}

variable "azure_aks_tenant_id" {
  description = "Azure AKS tenant ID"
}

variable "azure_aks_client_id" {
  description = "Azure AKS client ID"
}

variable "azure_aks_client_secret" {
  description = "Azure AKS client secret"
}

variable "install_count" {
  default = 5
  type = number
}
