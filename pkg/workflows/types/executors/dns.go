package executors

type ProvisionDNSDelegationRequest struct {
	InstallID string
	OrgID     string
	AppID     string
}

func (d ProvisionDNSDelegationRequest) Validate() error {
	return nil
}

type ProvisionDNSDelegationResponse struct{}

type DeprovisionDNSDelegationRequest struct{}

type DeprovisionDNSDelegationResponse struct{}
