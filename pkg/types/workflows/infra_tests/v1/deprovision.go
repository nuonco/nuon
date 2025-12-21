package infratestsv1

type DeprovisionRequest struct {
	CanaryID    string `json:"canary_id"`
	SandboxMode bool   `json:"sandbox_mode"`
}

func (req *DeprovisionRequest) Validate() error {
	return nil
}

type DeprovisionResponse struct {
	CanaryID string `json:"canary_id"`
}

func (req *DeprovisionResponse) Validate() error {
	return nil
}
