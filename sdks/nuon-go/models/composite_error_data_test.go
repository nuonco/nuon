package models

import (
	"encoding/json"
	"testing"
)

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

	t.Run("embedded in deploy", func(t *testing.T) {
		payload := `{"id": "dpl123", "composite_error": ` + awsPermissionErr + `}`
		var d AppInstallDeploy
		if err := json.Unmarshal([]byte(payload), &d); err != nil {
			t.Fatalf("deploy carrying a composite error must decode: %v", err)
		}
	})
}
