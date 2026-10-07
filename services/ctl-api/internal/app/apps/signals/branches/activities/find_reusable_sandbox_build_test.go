package activities

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestSandboxConfigsEquivalent(t *testing.T) {
	region := "us-west-2"
	current := &app.AppSandboxConfig{
		TerraformVersion: "1.9.0",
		Variables:        pgtype.Hstore{"region": &region},
		VariablesFiles:   pq.StringArray{"terraform.tfvars"},
		PublicGitVCSConfig: &app.PublicGitVCSConfig{
			Repo:      "nuonco/aws-eks-sandbox",
			Directory: ".",
			Branch:    "main",
		},
	}
	same := &app.AppSandboxConfig{
		TerraformVersion: "1.9.0",
		Variables:        pgtype.Hstore{"region": &region},
		VariablesFiles:   pq.StringArray{"terraform.tfvars"},
		PublicGitVCSConfig: &app.PublicGitVCSConfig{
			Repo:      "https://github.com/nuonco/aws-eks-sandbox.git",
			Directory: ".",
			Branch:    "main",
		},
	}
	otherVersion := *same
	otherVersion.TerraformVersion = "1.5.0"

	require.True(t, sandboxConfigsEquivalent(current, same))
	require.False(t, sandboxConfigsEquivalent(current, &otherVersion))
}
