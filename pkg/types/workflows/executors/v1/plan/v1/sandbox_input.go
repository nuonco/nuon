package planv1

func (s SandboxInputType) ToRunType() TerraformRunType {
	switch s {
	case SandboxInputType_SANDBOX_INPUT_TYPE_DEPROVISION:
		return TerraformRunType_TERRAFORM_RUN_TYPE_DESTROY
	case SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION:
		return TerraformRunType_TERRAFORM_RUN_TYPE_APPLY
	default:
		return TerraformRunType_TERRAFORM_RUN_TYPE_UNSPECIFIED
	}
}
