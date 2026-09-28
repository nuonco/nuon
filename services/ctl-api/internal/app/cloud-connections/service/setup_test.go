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
				assert.Contains(t, string(policy), "ec2:RunInstances")
				assert.Contains(t, string(policy), "ssm:GetParameters")
				for _, action := range []string{
					"autoscaling:DescribeLifecycleHooks", "iam:GetAccountSummary", "iam:ListPoliciesGrantingServiceAccess", "lambda:GetAccountSettings", "servicequotas:GetServiceQuota",
					"autoscaling:CreateOrUpdateTags", "autoscaling:DeleteTags", "autoscaling:DescribeAutoScalingInstances", "autoscaling:DescribeTags", "autoscaling:ResumeProcesses", "autoscaling:SetDesiredCapacity", "autoscaling:SuspendProcesses", "autoscaling:TerminateInstanceInAutoScalingGroup",
					"cloudformation:DescribeStackResource",
					"ec2:CreateLaunchTemplateVersion", "ec2:DeleteTags", "ec2:ReplaceRoute", "ec2:ReplaceRouteTableAssociation", "ec2:UpdateSecurityGroupRuleDescriptionsEgress", "ec2:UpdateSecurityGroupRuleDescriptionsIngress",
					"iam:CreatePolicyVersion", "iam:DeletePolicyVersion", "iam:DeleteRolePermissionsBoundary", "iam:ListEntitiesForPolicy", "iam:PutRolePermissionsBoundary", "iam:UpdateRole", "iam:UpdateRoleDescription",
					"lambda:GetFunctionCodeSigningConfig", "lambda:GetFunctionRecursionConfig", "lambda:GetFunctionScalingConfig", "lambda:GetRuntimeManagementConfig", "lambda:InvokeFunction",
					"logs:DeleteRetentionPolicy", "logs:DescribeIndexPolicies", "logs:DescribeResourcePolicies", "logs:GetDataProtectionPolicy", "logs:ListTagsForResource",
				} {
					assert.Contains(t, string(policy), action)
				}
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
