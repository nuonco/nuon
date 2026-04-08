package aws

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// BuildIIDAuthRequest creates a RunnerAuthAWSIIDRequest from an IID result
// and a runner ID.
func BuildIIDAuthRequest(iid *IIDResult, runnerID string) *models.ServiceRunnerAuthAWSIIDRequest {
	return &models.ServiceRunnerAuthAWSIIDRequest{
		Document:  &iid.Document,
		Signature: &iid.Signature,
		RunnerID:  &runnerID,
	}
}
