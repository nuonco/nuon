package aws

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	awstypes "github.com/nuonco/nuon/pkg/types/aws"
)

func BuildAuthRequest(stsRequest, tagsRequest *awstypes.PresignedRequest) *models.ServiceRunnerAuthAWSRequest {
	return &models.ServiceRunnerAuthAWSRequest{
		Sts: &models.AwsPresignedRequest{
			Method:  stsRequest.Method,
			URL:     stsRequest.URL,
			Headers: stsRequest.Headers,
		},
		Tags: &models.AwsPresignedRequest{
			Method:  tagsRequest.Method,
			URL:     tagsRequest.URL,
			Headers: tagsRequest.Headers,
		},
	}
}
