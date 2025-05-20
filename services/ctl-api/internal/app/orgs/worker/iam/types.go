package orgiam

type ProvisionIAMRequest struct {
	OrgID       string `json:"org_id"`
	Reprovision bool   `json:"reprovision"`

	WorkflowID string `json:"workflow_id"`
}

type ProvisionIAMResponse struct {
	RunnerRoleArn string
}

type DeprovisionIAMRequest struct {
	OrgID string

	WorkflowID string `json:"workflow_id"`
}

type DeprovisionIAMResponse struct{}
