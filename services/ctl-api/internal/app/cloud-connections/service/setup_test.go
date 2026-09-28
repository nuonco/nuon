package service

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func TestAWSSetup(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer, err := oidcissuer.New("https://api.example.com", key, "example-key")
	require.NoError(t, err)
	svc := &service{issuer: issuer}

	tests := map[string]struct {
		preset  app.CloudConnectionPreset
		want    []string
		notWant []string
	}{
		"stacks": {
			preset:  app.CloudConnectionPresetStacks,
			want:    []string{"cloudformation:CreateStack", "cloudformation:DescribeStacks", "iam:CreatePolicy", "iam:GetInstanceProfile", "iam:CreateServiceLinkedRole", "secretsmanager:CreateSecret"},
			notWant: []string{"ecr:"},
		},
		"custom trust only": {
			preset:  app.CloudConnectionPresetCustom,
			want:    []string{"sts:AssumeRoleWithWebIdentity", "api.example.com:sub", "org:org_example:connection:cc_example"},
			notWant: []string{"cloudformation:", "ecr:", "put-role-policy", "aws_iam_role_policy", "Policies:"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			connection := &app.CloudConnection{ID: "cc_example", OrgID: "org_example", Platform: app.CloudPlatformAWS, TargetID: "123456789012", DefaultRegion: "us-east-1", Preset: test.preset}
			got := svc.setup(connection)
			policy, err := json.Marshal(got.PermissionsPolicy)
			require.NoError(t, err)
			if test.preset == app.CloudConnectionPresetStacks {
				assert.Contains(t, string(policy), "cloudformation:DescribeStacks")
				assert.Contains(t, string(policy), "cloudformation:ListStacks")
				assert.Contains(t, string(policy), "ssm:GetParameters")
			} else {
				assert.Nil(t, got.PermissionsPolicy)
			}
			material := string(policy) + got.CLI + got.Terraform + got.CloudFormation
			for _, value := range test.want {
				assert.Contains(t, material, value)
			}
			for _, value := range test.notWant {
				assert.NotContains(t, material, value)
			}
			assert.NotContains(t, got.CLI, "thumbprint")
		})
	}
}
