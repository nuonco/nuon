package infratestsv1

type ProvisionRequest struct {
	CanaryID string `json:"canary_id"`
	OrgID    string `json:"org_id"` // should be generated internally

	TerraformVersion string `json:"terraform_version"`
	SandboxName      string `json:"sandbox_name"`
	SandboxRepo      string `json:"sandbox_repo"`
	SandboxBranch    string `json:"sandbox_branch"`

	// TODO(fd): never gonna use this in sandbox mode - remove it
	SandboxMode bool `json:"sandbox_mode"`
}

func (req *ProvisionRequest) Validate() error {
	return nil
}

type ProvisionResponse struct {
	CanaryID string `json:"canary_id"`
	OrgID    string `json:"org_id"`
}

func (req *ProvisionResponse) Validate() error {
	return nil
}
