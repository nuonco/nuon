package planv1

func (s SandboxInputType) ToRunType() TerraformRunType {
	switch s {
	// deprovision
	case SandboxInputType_SANDBOX_INPUT_TYPE_DEPROVISION:
		return TerraformRunType_TERRAFORM_RUN_TYPE_DESTROY

	// provision
	case SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION:
		return TerraformRunType_TERRAFORM_RUN_TYPE_APPLY
	case SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION_PLAN:
		return TerraformRunType_TERRAFORM_RUN_TYPE_PLAN
	default:
		return TerraformRunType_TERRAFORM_RUN_TYPE_UNSPECIFIED
	}
}
