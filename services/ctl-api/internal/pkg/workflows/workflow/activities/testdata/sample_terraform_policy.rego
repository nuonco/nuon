# Sample OPA policy for Terraform plan validation
# Tests against the fake_terraform_plan_display_contents.json sandbox plan

package nuon

# Deny KMS keys without key rotation enabled
deny contains msg if {
	resource := input.plan.resource_changes[_]
	resource.type == "aws_kms_key"
	resource.mode == "managed"
	resource.change.after.enable_key_rotation == false
	msg := sprintf("KMS key %s does not have key rotation enabled", [resource.address])
}

# Deny overly permissive security group rules (protocol -1 means all)
deny contains msg if {
	resource := input.plan.resource_changes[_]
	resource.type == "aws_security_group_rule"
	resource.mode == "managed"
	resource.change.after.protocol == "-1"
	resource.change.after.from_port == 0
	resource.change.after.to_port == 0
	msg := sprintf("Security group rule %s allows all traffic (protocol: -1, ports: 0-0)", [resource.address])
}

# Warn on resources missing Environment tag
warn contains msg if {
	resource := input.plan.resource_changes[_]
	resource.mode == "managed"
	resource.change.after.tags_all
	not resource.change.after.tags_all.Environment
	msg := sprintf("Resource %s is missing Environment tag", [resource.address])
}
