package ecrrepository

type ProvisionECRRepositoryRequest struct {
	OrgID string
	AppID string

	WorkflowID string
}

type ProvisionECRRepositoryResponse struct {
	RegistryID     string
	RepositoryName string
	RepositoryARN  string
	RepositoryURI  string
}
