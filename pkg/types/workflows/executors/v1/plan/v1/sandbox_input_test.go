package planv1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandboxInputType_ToRunType(t *testing.T) {
	assert.Equal(t,
		SandboxInputType_SANDBOX_INPUT_TYPE_DEPROVISION.ToRunType(),
		TerraformRunType_TERRAFORM_RUN_TYPE_DESTROY)

	assert.Equal(t,
		SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION.ToRunType(),
		TerraformRunType_TERRAFORM_RUN_TYPE_APPLY)

	assert.Equal(t,
		SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION_PLAN.ToRunType(),
		TerraformRunType_TERRAFORM_RUN_TYPE_PLAN)
}
