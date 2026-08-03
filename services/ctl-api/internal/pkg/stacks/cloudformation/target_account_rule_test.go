package cloudformation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func TestTemplate_TargetAccountRule(t *testing.T) {
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

	tpl := &Templates{cfg: &internal.Config{UseLocalRunners: true}}

	t.Run("no rules section without a target account", func(t *testing.T) {
		tmplJSON, checksum, err := tpl.Template(newInput())
		require.NoError(t, err)
		require.NotEmpty(t, checksum)

		var tmpl map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(tmplJSON, &tmpl))
		assert.NotContains(t, tmpl, "Rules")
	})

	t.Run("target account pins the template via a rules assertion", func(t *testing.T) {
		inp := newInput()
		inp.TargetAWSAccountID = "123456789012"

		tmplJSON, _, err := tpl.Template(inp)
		require.NoError(t, err)

		var tmpl struct {
			Resources map[string]json.RawMessage `json:"Resources"`
			Rules     map[string]struct {
				Assertions []struct {
					Assert struct {
						FnEquals []json.RawMessage `json:"Fn::Equals"`
					} `json:"Assert"`
					AssertDescription string `json:"AssertDescription"`
				} `json:"Assertions"`
			} `json:"Rules"`
		}
		require.NoError(t, json.Unmarshal(tmplJSON, &tmpl))

		require.Contains(t, tmpl.Rules, "NuonTargetAWSAccount")
		assertions := tmpl.Rules["NuonTargetAWSAccount"].Assertions
		require.Len(t, assertions, 1)

		equals := assertions[0].Assert.FnEquals
		require.Len(t, equals, 2)
		assert.JSONEq(t, `{"Ref": "AWS::AccountId"}`, string(equals[0]))
		assert.JSONEq(t, `"123456789012"`, string(equals[1]))

		assert.Contains(t, assertions[0].AssertDescription, "123456789012")
		assert.Contains(t, assertions[0].AssertDescription, "inl123")

		// The injection must not disturb the rest of the template.
		assert.Contains(t, tmpl.Resources, "RunnerSecurityGroup")
	})

	t.Run("checksum covers the rule", func(t *testing.T) {
		_, without, err := tpl.Template(newInput())
		require.NoError(t, err)

		inp := newInput()
		inp.TargetAWSAccountID = "123456789012"
		_, with, err := tpl.Template(inp)
		require.NoError(t, err)

		assert.NotEqual(t, without, with)
	})
}
