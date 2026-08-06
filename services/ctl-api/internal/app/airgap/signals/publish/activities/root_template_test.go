package activities

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func rootTemplateFixture(script string) []byte {
	doc := map[string]any{
		"Parameters": map[string]any{
			"PhoneHomeS3Bucket": map[string]any{"Type": "String", "Default": ""},
			"PhoneHomeS3Key":    map[string]any{"Type": "String", "Default": ""},
		},
		"Resources": map[string]any{
			"RunnerPhoneHome": map[string]any{
				"Type": "AWS::Lambda::Function",
				"Properties": map[string]any{
					"Code": map[string]any{"ZipFile": script},
					"Environment": map[string]any{
						"Variables": map[string]any{
							"NUON_PHONE_HOME_SECRET_ARN":    "arn:aws:secretsmanager:us-west-2:111111111111:secret:nuon/phone-home/inl123",
							"NUON_PHONE_HOME_SECRET_REGION": "us-west-2",
						},
					},
				},
			},
			"RunnerPhoneHomeRole": map[string]any{
				"Type": "AWS::IAM::Role",
				"Properties": map[string]any{
					"Policies": []any{
						map[string]any{"PolicyName": "PhoneHomeSecretPolicy"},
						map[string]any{"PolicyName": "CloudWatchLogsPolicy"},
					},
				},
			},
			"PhoneHomeProps": map[string]any{
				"Type": "AWS::CloudFormation::CustomResource",
				"Properties": map[string]any{
					"url":             "https://api.vendor.example/v1/installs/inl123/phone-home/ph123",
					"phone_home_type": "aws",
				},
			},
			"RunnerAutoScalingGroup": map[string]any{
				"Type": "AWS::CloudFormation::Stack",
				"Properties": map[string]any{
					"Parameters": map[string]any{
						"RunnerApiToken": "nuon-live-token",
						"RunnerApiUrl":   "https://runner.vendor.example",
						"InstallId":      "inl123",
					},
					"Tags": []any{
						map[string]any{"Key": "nuon_runner_api_token", "Value": "nuon-live-token"},
						map[string]any{"Key": "nuon_runner_api_url", "Value": "https://runner.vendor.example"},
						map[string]any{"Key": "nuon_install_id", "Value": "inl123"},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(doc)
	return raw
}

func TestPrepareRootTemplateForBundleRejectsUnrenderedNuonTemplates(t *testing.T) {
	doc := rootTemplateFixture("import os\nNUON_PHONE_HOME_S3_BUCKET")
	tainted := strings.Replace(string(doc), "inl123", "{{ .nuon.install.id }}", 1)

	_, err := prepareRootTemplateForBundle([]byte(tainted))
	require.ErrorContains(t, err, "unrendered template expression")
	require.ErrorContains(t, err, ".nuon.install.id")
}

func TestPrepareRootTemplateForBundleRejectsInputPlaceholders(t *testing.T) {
	doc := rootTemplateFixture("import os\nNUON_PHONE_HOME_S3_BUCKET")
	tainted := strings.Replace(string(doc), "inl123", "__NUON_INPUT_env__", 1)

	_, err := prepareRootTemplateForBundle([]byte(tainted))
	require.ErrorContains(t, err, "references install inputs")
}

func TestPrepareRootTemplateForBundleRemovesRunnerAPISurfaces(t *testing.T) {
	out, err := prepareRootTemplateForBundle(rootTemplateFixture("import os\nNUON_PHONE_HOME_S3_BUCKET"))
	require.NoError(t, err)
	require.NotContains(t, string(out), "nuon-live-token")
	require.NotContains(t, string(out), "RunnerApiToken")
	require.NotContains(t, string(out), "RunnerApiUrl")
	require.NotContains(t, string(out), "nuon_runner_api_token")
	require.NotContains(t, string(out), "nuon_runner_api_url")
	require.NotContains(t, string(out), "runner.vendor.example")

	var doc map[string]any
	require.NoError(t, json.Unmarshal(out, &doc))
	resources := doc["Resources"].(map[string]any)
	props := resources["RunnerAutoScalingGroup"].(map[string]any)["Properties"].(map[string]any)
	require.Equal(t, "inl123", props["Parameters"].(map[string]any)["InstallId"])
	tags := props["Tags"].([]any)
	require.Len(t, tags, 1)
	require.Equal(t, "nuon_install_id", tags[0].(map[string]any)["Key"])
}

func TestPrepareRootTemplateForBundleRemovesPhoneHomeControlPlane(t *testing.T) {
	out, err := prepareRootTemplateForBundle(rootTemplateFixture("import os\nNUON_PHONE_HOME_S3_BUCKET"))
	require.NoError(t, err)
	require.NotContains(t, string(out), "api.vendor.example")
	require.NotContains(t, string(out), "NUON_PHONE_HOME_SECRET_ARN")
	require.NotContains(t, string(out), "PhoneHomeSecretPolicy")

	var doc map[string]any
	require.NoError(t, json.Unmarshal(out, &doc))
	resources := doc["Resources"].(map[string]any)

	phoneHomeProps := resources["PhoneHomeProps"].(map[string]any)["Properties"].(map[string]any)
	require.Equal(t, "", phoneHomeProps["url"])
	require.Equal(t, "aws", phoneHomeProps["phone_home_type"])

	lambdaEnv := resources["RunnerPhoneHome"].(map[string]any)["Properties"].(map[string]any)["Environment"].(map[string]any)["Variables"].(map[string]any)
	require.Empty(t, lambdaEnv)

	policies := resources["RunnerPhoneHomeRole"].(map[string]any)["Properties"].(map[string]any)["Policies"].([]any)
	require.Len(t, policies, 1)
	require.Equal(t, "CloudWatchLogsPolicy", policies[0].(map[string]any)["PolicyName"])
}

func TestPrepareRootTemplateForBundleRequiresS3Parameter(t *testing.T) {
	raw := rootTemplateFixture("NUON_PHONE_HOME_S3_BUCKET")
	var doc map[string]any
	require.NoError(t, json.Unmarshal(raw, &doc))
	delete(doc["Parameters"].(map[string]any), "PhoneHomeS3Bucket")
	raw, _ = json.Marshal(doc)

	_, err := prepareRootTemplateForBundle(raw)
	require.ErrorContains(t, err, "PhoneHomeS3Bucket")
}

func TestPrepareRootTemplateForBundleRequiresS3CapableScript(t *testing.T) {
	_, err := prepareRootTemplateForBundle(rootTemplateFixture("legacy script, posts to control plane only"))
	require.ErrorContains(t, err, "S3 rendezvous")

	var doc map[string]any
	require.NoError(t, json.Unmarshal(rootTemplateFixture("x"), &doc))
	code := doc["Resources"].(map[string]any)["RunnerPhoneHome"].(map[string]any)["Properties"].(map[string]any)["Code"].(map[string]any)
	delete(code, "ZipFile")
	raw, _ := json.Marshal(doc)
	_, err = prepareRootTemplateForBundle(raw)
	require.ErrorContains(t, err, "inline RunnerPhoneHome Lambda source")
}

func TestPrepareRootTemplateForBundleRejectsInvalidJSON(t *testing.T) {
	_, err := prepareRootTemplateForBundle([]byte("not json"))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "decode install stack template"))
}

func TestCompileRootTemplateWithoutInstall(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/phonehome.py", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("import os\nNUON_PHONE_HOME_S3_BUCKET = os.environ.get('NUON_PHONE_HOME_S3_BUCKET')"))
	})
	mux.HandleFunc("/vpc.yaml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`
Parameters:
  NuonInstallID:
    Type: String
  NuonAppID:
    Type: String
  NuonOrgID:
    Type: String
Outputs:
  VPC: {}
  RunnerSubnet: {}
  PublicSubnets: {}
  PrivateSubnets: {}
`))
	})
	mux.HandleFunc("/runner.yaml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`
Parameters:
  RunnerApiToken:
    Type: String
  RunnerApiUrl:
    Type: String
  RunnerEnvVars:
    Type: String
Outputs:
  RunnerInstanceRole: {}
`))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	ctlCfg := &internal.Config{
		PublicAPIURL:            "https://api.vendor.example",
		RunnerAPIURL:            "https://runner.vendor.example",
		RunnerContainerImageTag: "v1.2.3",
	}
	appCfg := &app.AppConfig{
		ID:    "app-config",
		OrgID: "org-test",
		AppID: "app-test",
		RunnerConfig: app.AppRunnerConfig{
			Type:               app.AppRunnerTypeAWS,
			PhoneHomeScriptURL: server.URL + "/phonehome.py",
		},
		StackConfig: app.AppStackConfig{
			Name:                    "test-stack",
			VPCNestedTemplateURL:    server.URL + "/vpc.yaml",
			RunnerNestedTemplateURL: server.URL + "/runner.yaml",
		},
		PermissionsConfig: app.AppPermissionsConfig{
			Roles: []app.AppAWSIAMRoleConfig{{
				Type:                         app.AWSIAMRoleTypeRunnerProvision,
				Name:                         "{{.nuon.install.id}}-provision",
				CloudFormationStackName:      "RunnerProvision",
				CloudFormationStackParamName: "EnableRunnerProvision",
				PermissionsBoundaryJSON:      []byte(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`),
			}},
		},
	}

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE app_input_configs (
		id text primary key, created_at datetime, deleted_at integer not null default 0,
		app_config_id text
	)`).Error)

	contents, source, err := (&Activities{cfg: ctlCfg, db: db}).compileRootTemplate(context.Background(), appCfg, "inl-synthetic")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(source, "compiled:cloudformation:"))
	require.NotContains(t, string(contents), "RunnerApiToken")
	require.NotContains(t, string(contents), "nuon_runner_api_token")
	require.NotContains(t, string(contents), "runner.vendor.example")
	require.Contains(t, string(contents), "PhoneHomeS3Bucket")
	require.Contains(t, string(contents), "NUON_PHONE_HOME_S3_BUCKET")
	require.Contains(t, string(contents), "inl-synthetic")
	require.Contains(t, string(contents), "inl-synthetic-provision")
	require.NotContains(t, string(contents), "{{")
}

func TestValidateRunnerNestedTemplateAirgapCompatible(t *testing.T) {
	require.NoError(t, validateRunnerNestedTemplateAirgapCompatible([]byte(`
Parameters:
  RunnerApiUrl:
    Type: String
    Default: https://runner.nuon.co
`)))

	require.NoError(t, validateRunnerNestedTemplateAirgapCompatible([]byte(`
Parameters:
  RunnerApiToken:
    Type: String
    Default: ""
`)))

	err := validateRunnerNestedTemplateAirgapCompatible([]byte(`
Parameters:
  RunnerApiToken:
    Type: String
`))
	require.ErrorContains(t, err, "RunnerApiToken")
	require.ErrorContains(t, err, "airgap-compatible")
}
