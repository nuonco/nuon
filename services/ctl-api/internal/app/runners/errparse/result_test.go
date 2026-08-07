package errparse_test

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/all"
)

func TestParseRunnerJobResultProviderGate(t *testing.T) {
	runnerJob := &app.RunnerJob{
		Type:      app.RunnerJobTypeTerraformDeploy,
		OwnerType: "install_deploys",
		OwnerID:   "dpl123",
	}
	metadata := map[string]string{
		errparse.ErrorMetadataOutput: "Error: creating S3 Bucket: AccessDenied: User: arn:aws:iam::123:role/runner is not authorized to perform: s3:CreateBucket on resource: arn:aws:s3:::bucket",
	}

	tests := []struct {
		name     string
		provider errparse.Provider
		wantType string
	}{
		{name: "AWS parser applies to AWS jobs", provider: errparse.ProviderAWS, wantType: "terraform.aws_permission"},
		{name: "AWS parser does not apply to Azure jobs", provider: errparse.ProviderAzure, wantType: "terraform.error"},
		{name: "unknown provider fails open", provider: errparse.ProviderUnknown, wantType: "terraform.aws_permission"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := errparse.ParseRunnerJobResult(false, metadata, runnerJob, func() errparse.Provider {
				return test.provider
			})
			if err != nil {
				t.Fatalf("parse result: %v", err)
			}
			if got == nil || string(got.Type) != test.wantType {
				t.Fatalf("type = %v, want %q", got, test.wantType)
			}
		})
	}
}

func TestParseRunnerJobResultResolvesProviderLazily(t *testing.T) {
	resolved := false
	runnerJob := &app.RunnerJob{Type: app.RunnerJobTypeTerraformDeploy}
	got, err := errparse.ParseRunnerJobResult(false, map[string]string{
		errparse.ErrorMetadataMessage: "some unrelated failure",
	}, runnerJob, func() errparse.Provider {
		resolved = true
		return errparse.ProviderAWS
	})
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}
	if got == nil || got.Type != "generic" {
		t.Fatalf("expected generic error, got %+v", got)
	}
	if resolved {
		t.Fatal("provider was resolved without a provider-specific signal")
	}
}

func TestResolveRunnerJobProvider(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.Exec(`CREATE TABLE runners (id text PRIMARY KEY, runner_group_id text, deleted_at integer DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create runners table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE runner_groups (id text PRIMARY KEY, platform text, deleted_at integer DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create runner groups table: %v", err)
	}

	tests := []struct {
		name     string
		runnerID string
		groupID  string
		platform app.AppRunnerType
		want     errparse.Provider
	}{
		{name: "AWS", runnerID: "runner-aws", groupID: "group-aws", platform: app.AppRunnerTypeAWS, want: errparse.ProviderAWS},
		{name: "Azure", runnerID: "runner-azure", groupID: "group-azure", platform: app.AppRunnerTypeAzure, want: errparse.ProviderAzure},
		{name: "GCP", runnerID: "runner-gcp", groupID: "group-gcp", platform: app.AppRunnerTypeGCP, want: errparse.ProviderGCP},
		{name: "local is unknown", runnerID: "runner-local", groupID: "group-local", platform: app.AppRunnerTypeLocal, want: errparse.ProviderUnknown},
	}
	for _, test := range tests {
		if err := db.Exec(`INSERT INTO runner_groups (id, platform, deleted_at) VALUES (?, ?, 0)`, test.groupID, test.platform).Error; err != nil {
			t.Fatalf("insert runner group: %v", err)
		}
		if err := db.Exec(`INSERT INTO runners (id, runner_group_id, deleted_at) VALUES (?, ?, 0)`, test.runnerID, test.groupID).Error; err != nil {
			t.Fatalf("insert runner: %v", err)
		}
		t.Run(test.name, func(t *testing.T) {
			got := errparse.ResolveRunnerJobProvider(context.Background(), db, &app.RunnerJob{RunnerID: test.runnerID})
			if got != test.want {
				t.Fatalf("provider = %q, want %q", got, test.want)
			}
		})
	}

	if got := errparse.ResolveRunnerJobProvider(context.Background(), db, &app.RunnerJob{RunnerID: "missing"}); got != errparse.ProviderUnknown {
		t.Fatalf("missing runner provider = %q, want unknown", got)
	}
}
