package ecrrepository

import "fmt"

// BuildGARResponse synthesizes a ProvisionECRRepositoryResponse for a GCP-hosted
// control plane. GCP Artifact Registry Docker repositories accept arbitrary path
// hierarchies, so apps share the management repository with org/app as a path
// prefix rather than getting their own GCP-side repo.
func BuildGARResponse(repositoryURL, orgID, appID, region string) *ProvisionECRRepositoryResponse {
	return &ProvisionECRRepositoryResponse{
		RepositoryName: fmt.Sprintf("%s/%s", orgID, appID),
		RepositoryURI:  fmt.Sprintf("%s/%s/%s", repositoryURL, orgID, appID),
		Region:         region,
	}
}

// BuildACRResponse synthesizes a ProvisionECRRepositoryResponse for an
// Azure-hosted control plane. ACR uses the same shared-registry pattern as GAR.
func BuildACRResponse(registryURL, orgID, appID, region string) *ProvisionECRRepositoryResponse {
	return &ProvisionECRRepositoryResponse{
		RepositoryName: fmt.Sprintf("%s/%s", orgID, appID),
		RepositoryURI:  fmt.Sprintf("%s/%s/%s", registryURL, orgID, appID),
		Region:         region,
	}
}
