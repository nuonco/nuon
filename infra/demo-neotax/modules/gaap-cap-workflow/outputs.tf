output "step_0_coordinator_arn" {
  value = aws_sqs_queue.step_0_coordinator.arn
}

output "step_0_children_arn" {
  value = aws_sqs_queue.step_0_children.arn
}

output "step_1_coordinator_arn" {
  value = aws_sqs_queue.step_1_coordinator.arn
}
