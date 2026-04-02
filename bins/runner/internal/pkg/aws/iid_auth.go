package aws

import (
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

// BuildIIDAuthRequest creates a RunnerAuthAWSIIDRequest from an IID result
// and a runner ID.
func BuildIIDAuthRequest(iid *IIDResult, runnerID string) *nuonrunner.RunnerAuthAWSIIDRequest {
	return &nuonrunner.RunnerAuthAWSIIDRequest{
		Document:  iid.Document,
		Signature: iid.Signature,
		RunnerID:  runnerID,
	}
}
