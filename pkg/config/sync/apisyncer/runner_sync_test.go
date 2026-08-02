package apisyncer

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

// Fields that deliberately do not reach the API.
var runnerFieldsNotSynced = map[string]string{
	"Source":    "a local file pointer, resolved during parsing",
	"EnvVars":   "deprecated; EnvVarMap is what ships",
	"EnvVarMap": "carried as a map, so its value is not a scalar sentinel",
}

// A new field on config.AppRunnerConfig that is added to the SDK model but never
// assigned in runnerConfigRequest fails completely silently — the CLI parses it, the
// API would accept it, and the value never leaves the machine. That is how
// phone_home_script_url was lost.
//
// So this walks the config struct rather than asserting on a fixed list: add a field
// and this fails until it is either mapped or explicitly excused above.
func TestRunnerConfigRequestCarriesEveryField(t *testing.T) {
	cfgType := reflect.TypeOf(config.AppRunnerConfig{})

	cfg := &config.AppRunnerConfig{}
	v := reflect.ValueOf(cfg).Elem()

	// A distinct sentinel per string field, so a mapping that assigns the wrong
	// source field is caught too, not just a missing one.
	sentinels := map[string]string{}
	for i := 0; i < cfgType.NumField(); i++ {
		f := cfgType.Field(i)
		if _, skip := runnerFieldsNotSynced[f.Name]; skip {
			continue
		}
		if f.Type.Kind() != reflect.String {
			continue
		}
		sentinel := "sentinel-" + strings.ToLower(f.Name)
		v.Field(i).SetString(sentinel)
		sentinels[f.Name] = sentinel
	}
	require.NotEmpty(t, sentinels, "reflection found no string fields to check")

	req := runnerConfigRequest("appcfg_1", cfg)
	byts, err := json.Marshal(req)
	require.NoError(t, err)
	payload := string(byts)

	for field, sentinel := range sentinels {
		assert.Contains(t, payload, sentinel,
			"config.AppRunnerConfig.%s never reaches the API — map it in runnerConfigRequest "+
				"or excuse it in runnerFieldsNotSynced", field)
	}

	assert.Contains(t, payload, "appcfg_1", "the app config ID must be carried")
}

// The regression itself, named so a failure points straight at the symptom.
func TestRunnerConfigRequestCarriesPhoneHomeScriptURL(t *testing.T) {
	const url = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/fd/ph-auth/scripts/aws/phonehome.py"

	req := runnerConfigRequest("appcfg_1", &config.AppRunnerConfig{
		RunnerType:         "aws",
		PhoneHomeScriptURL: url,
	})

	assert.Equal(t, url, req.PhoneHomeScriptURL)
}

// EnvVarMap is excused from the sentinel walk, so cover it explicitly rather than
// leaving it untested.
func TestRunnerConfigRequestCarriesEnvVars(t *testing.T) {
	req := runnerConfigRequest("appcfg_1", &config.AppRunnerConfig{
		RunnerType: "aws",
		EnvVarMap:  map[string]string{"DEBUG": "true"},
	})

	assert.Equal(t, map[string]string{"DEBUG": "true"}, req.EnvVars)
}
