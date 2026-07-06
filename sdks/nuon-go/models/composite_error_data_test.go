package models

import (
	"encoding/json"
	"testing"
)

// A composite error's `data` is a polymorphic, per-error-type object (e.g. the
// marshaled AWSPermissionError). It must decode into the generated model as an
// arbitrary object, not a typed array. Regression guard for the "failed to
// fetch deploys" outage, where `data` was generated as []int64 and every deploy
// carrying a composite error failed to unmarshal in the dashboard BFF.
func TestCompositeErrorDataDecodesObjectPayload(t *testing.T) {
	const awsPermissionErr = `{
		"version": 1,
		"type": "terraform.aws_permission",
		"severity": "error",
		"message": "Missing AWS IAM permission: s3:CreateBucket",
		"sections": [{"heading": "Why", "body": "denied"}],
		"data": {
			"action": "s3:CreateBucket",
			"resource": "arn:aws:s3:::acme-prod-assets",
			"principal": "arn:aws:iam::123:role/nuon-runner",
			"aws_error_code": "AccessDenied"
		},
		"hints": {"skip_auto_retry": "true"}
	}`

	t.Run("standalone model", func(t *testing.T) {
		var ced CompositeerrorsCompositeErrorData
		if err := json.Unmarshal([]byte(awsPermissionErr), &ced); err != nil {
			t.Fatalf("composite error data with object payload must decode: %v", err)
		}
		if _, ok := ced.Data.(map[string]any); !ok {
			t.Fatalf("data should decode as an object, got %T", ced.Data)
		}
	})

	// The install deploy list is the surface that broke: a single deploy with a
	// composite error poisoned the whole response decode in the BFF.
	t.Run("embedded in deploy", func(t *testing.T) {
		payload := `{"id": "dpl123", "composite_error": ` + awsPermissionErr + `}`
		var d AppInstallDeploy
		if err := json.Unmarshal([]byte(payload), &d); err != nil {
			t.Fatalf("deploy carrying a composite error must decode: %v", err)
		}
	})
}
