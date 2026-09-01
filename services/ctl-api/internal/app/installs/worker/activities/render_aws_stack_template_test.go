package activities

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

const mockOutputTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Resources:
  Bucket:
    Type: Custom::Bucket
Outputs:
  BucketArn:
    Value: !Ref Bucket
  BucketName:
    Value: !Ref Bucket
`

func TestRenderAWSStackTemplate_CustomStacksOutputMapAgreesWithSplit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockOutputTemplateYAML))
	}))
	defer server.Close()

	customStacks := []config.CustomNestedStack{
		{Name: "my-bucket", TemplateURL: server.URL, Index: 0},
	}

	inp := stacks.TemplateInput{
		Install: &app.Install{ID: "test-install-id", AppID: "test-app-id", OrgID: "test-org-id"},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{CustomNestedStacks: customStacks},
		},
		Settings:         &app.RunnerGroupSettings{},
		CustomStacksOnly: true,
	}

	a := &Activities{cfTemplates: cloudformation.NewTemplates(cloudformation.Params{Cfg: &internal.Config{}})}

	res, err := a.RenderAWSStackTemplate(context.Background(), &RenderAWSStackTemplateRequest{Input: inp})
	require.NoError(t, err)

	tmpl, _, err := a.cfTemplates.Template(&inp)
	require.NoError(t, err)

	flatOutputs := make(map[string]string, len(tmpl.Outputs))
	for name := range tmpl.Outputs {
		flatOutputs[name] = name
	}
	want := map[string]map[string]string{}
	for stackName, stackResult := range cloudformation.SplitCustomStacksOnlyOutputs(flatOutputs, []string{"my-bucket"}) {
		want[stackName] = stackResult["outputs"]
	}

	assert.Equal(t, want, res.CustomStacksOutputMap)
	assert.Equal(t, map[string]string{
		"BucketArn":  "MyBucketBucketArn",
		"BucketName": "MyBucketBucketName",
	}, res.CustomStacksOutputMap["my-bucket"])
}

func TestRenderAWSStackTemplate_CustomStacksOutputMapEmptyWhenNotCustomStacksOnly(t *testing.T) {
	a := &Activities{cfTemplates: cloudformation.NewTemplates(cloudformation.Params{Cfg: &internal.Config{}})}

	inp := stacks.TemplateInput{
		Install:          &app.Install{ID: "test-install-id", AppID: "test-app-id", OrgID: "test-org-id"},
		AppCfg:           &app.AppConfig{},
		Settings:         &app.RunnerGroupSettings{},
		CustomStacksOnly: false,
	}

	res, _ := a.RenderAWSStackTemplate(context.Background(), &RenderAWSStackTemplateRequest{Input: inp})
	assert.Nil(t, res.CustomStacksOutputMap)
}
