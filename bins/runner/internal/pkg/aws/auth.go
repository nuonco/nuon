package aws

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	awstypes "github.com/nuonco/nuon/pkg/types/aws"
)

// BuildBoundSTSAuthRequest creates a RunnerAuthAWSSTSRequest from a bound presigned STS request.
func BuildBoundSTSAuthRequest(authType, runnerID string, stsRequest *awstypes.PresignedRequest) *models.ServiceRunnerAuthAWSSTSRequest {
	return &models.ServiceRunnerAuthAWSSTSRequest{
		Type:     &authType,
		RunnerID: &runnerID,
		Sts: &models.AwsPresignedRequest{
			Method:  stsRequest.Method,
			URL:     stsRequest.URL,
			Headers: stsRequest.Headers,
		},
	}
}

// BuildAuthRequest creates a RunnerAuthAWSRequest from presigned STS and EC2 tags requests.
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
