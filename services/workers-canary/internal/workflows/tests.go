package workflows

import (
	"fmt"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) execTests(ctx workflow.Context,
	req *canaryv1.ProvisionRequest,
	outputs *activities.TerraformRunOutputs,
	orgID string,
	apiToken string,
) error {
	var testsResp activities.ListTestsResponse
	if err := w.defaultExecGetActivity(ctx, w.acts.ListTests, &activities.ListTestsRequest{}, &testsResp); err != nil {
		w.metricsWriter.Incr(ctx, "provision", 1, "status:error", "step:list_tests")
		return fmt.Errorf("unable to list tests: %w", err)
	}

	tfOutputsPath := fmt.Sprintf("/tmp/%s.json", req.CanaryId)
	env := map[string]string{
		"NUON_API_URL":    w.cfg.APIURL,
		"NUON_API_TOKEN":  apiToken,
		"NUON_ORG_ID":     orgID,
		"SANDBOX":         fmt.Sprintf("%t", req.SandboxMode),
		"TF_OUTPUTS_PATH": tfOutputsPath,
	}

	for idx, test := range testsResp.Tests {
		req := activities.ExecTestScriptRequest{
			Path:          test,
			Env:           env,
			TFOutputs:     outputs,
			TFOutputsPath: tfOutputsPath,
			InstallCLI:    idx == 0,
		}

		var testResp activities.ExecTestScriptResponse
		if err := w.defaultExecTestActivity(ctx, w.acts.ExecTestScript, req, &testResp); err != nil {
			w.metricsWriter.Incr(ctx, "provision", 1, "status:error", fmt.Sprintf("step:execute_test_%d", idx+1))
			return fmt.Errorf("unable to execute test: %w", err)
		}
		w.metricsWriter.Incr(ctx, "provision", 1, "status:ok", fmt.Sprintf("step:execute_test_%d", idx+1))
	}

	return nil
}
