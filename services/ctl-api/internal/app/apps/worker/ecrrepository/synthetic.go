package ecrrepository

import "fmt"

func BuildGARResponse(repositoryURL, orgID, appID, region string) *ProvisionECRRepositoryResponse {
	return &ProvisionECRRepositoryResponse{
		RepositoryName: fmt.Sprintf("%s/%s", orgID, appID),
		RepositoryURI:  fmt.Sprintf("%s/%s/%s", repositoryURL, orgID, appID),
		Region:         region,
	}
}

func BuildACRResponse(registryURL, orgID, appID, region string) *ProvisionECRRepositoryResponse {
	return &ProvisionECRRepositoryResponse{
		RepositoryName: fmt.Sprintf("%s/%s", orgID, appID),
		RepositoryURI:  fmt.Sprintf("%s/%s/%s", registryURL, orgID, appID),
		Region:         region,
	}
}
