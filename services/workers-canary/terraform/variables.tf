variable "aws_eks_iam_role_arn" {
  description = "IAM role for AWS EKS sandbox"
}

variable "aws_ecs_iam_role_arn" {
  description = "IAM role for AWS ECS sandbox"
}

variable "install_count" {
  default = 5
  type = number
}
