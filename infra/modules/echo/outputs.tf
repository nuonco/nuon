output "output" {
  description = "this is equal to the input value"
  value       = var.input
}

output "account_id" {
  description = "account id this is being used from"
  value       = data.aws_caller_identity.current.account_id
}
