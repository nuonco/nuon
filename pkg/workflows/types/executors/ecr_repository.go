package executors

const (
	ProvisionECRRepositoryWorkflowName   string = "ProvisionECRRepository"
	DeprovisionECRRepositoryWorkflowName string = "DeprovisionECRRepository"
)

type ProvisionECRRepositoryRequest struct {
	OrgID string
	AppID string
}

type ProvisionECRRepositoryResponse struct {
	RegistryID     string
	RepositoryName string
	RepositoryARN  string
	RepositoryURI  string
}

type DeprovisionECRRepositoryRequest struct  {
	OrgID string
	AppID string
}

type DeprovisionECRRepositoryResponse struct{}
