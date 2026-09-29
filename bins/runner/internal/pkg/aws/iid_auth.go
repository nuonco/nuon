package aws

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func BuildIIDAuthRequest(iid *IIDResult, runnerID string) *models.ServiceRunnerAuthAWSIIDRequest {
	return &models.ServiceRunnerAuthAWSIIDRequest{
		Document:  &iid.Document,
		Signature: &iid.Signature,
		RunnerID:  &runnerID,
	}
}
