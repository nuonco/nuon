package cloudformation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awslabs/goformation/v7/cloudformation/ec2"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func TestGetAWSTemplate_RunnerResources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(mockVPCTemplateYAML))
	}))
	defer server.Close()

	newInput := func() *stacks.TemplateInput {
		return &stacks.TemplateInput{
			Install: &app.Install{
				ID:    "inl123",
				AppID: "app123",
				OrgID: "org123",
			},
			CloudFormationStackVersion: &app.InstallStackVersion{PhoneHomeURL: server.URL + "/phone-home"},
			AppCfg: &app.AppConfig{
				StackConfig: app.AppStackConfig{
					VPCNestedTemplateURL:    server.URL + "/vpc.yaml",
					RunnerNestedTemplateURL: server.URL + "/runner.yaml",
				},
			},
			Runner:   &app.Runner{ID: "run123"},
			Settings: &app.RunnerGroupSettings{},
		}
	}

	// The sandbox terraform looks the runner security group up by tag to grant it
	// cluster access, so it has to exist even when no runner instance does.
	t.Run("runner security group is created with local runners", func(t *testing.T) {
		tpl := &Templates{cfg: &internal.Config{UseLocalRunners: true}}

		tmpl, err := tpl.getAWSTemplate(newInput())
		require.NoError(t, err)

		require.Contains(t, tmpl.Resources, "RunnerSecurityGroup")
		sg, ok := tmpl.Resources["RunnerSecurityGroup"].(*ec2.SecurityGroup)
		require.True(t, ok)
		assert.Contains(t, sg.Tags, tags.Tag{Key: "network.nuon.co/domain", Value: "runner"})

		assert.NotContains(t, tmpl.Resources, "RunnerAutoScalingGroup")
		assert.NotContains(t, tmpl.Resources, "RunnerCloudWatchLogGroup")
	})

	t.Run("runner asg and logs are created with cloud runners", func(t *testing.T) {
		tpl := &Templates{cfg: &internal.Config{UseLocalRunners: false}}

		tmpl, err := tpl.getAWSTemplate(newInput())
		require.NoError(t, err)

		assert.Contains(t, tmpl.Resources, "RunnerSecurityGroup")
		assert.Contains(t, tmpl.Resources, "RunnerAutoScalingGroup")
		assert.Contains(t, tmpl.Resources, "RunnerCloudWatchLogGroup")
		assert.Contains(t, tmpl.Resources, "RunnerCloudWatchLogStream")
		assert.Contains(t, tmpl.Resources, "RunnerCloudWatchLogPolicy")
	})
}
