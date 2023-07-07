output "step_0_coordinator_arn" {
  value = module.gaap_cap_workflow.step_0_coordinator_arn
}

output "step_0_children_arn" {
  value = module.gaap_cap_workflow.step_0_children_arn
}

output "step_1_coordinator_arn" {
  value = module.gaap_cap_workflow.step_1_coordinator_arn
}

output "db_instance_endpoint" {
  value = module.database.db_instance_endpoint
}
