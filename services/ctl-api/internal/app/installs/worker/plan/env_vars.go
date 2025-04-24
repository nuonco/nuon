package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (p *Planner) getEnvVars(ctx workflow.Context, run *app.InstallActionWorkflowRun) (map[string]string, error) {
	token, err := activities.AwaitCreateActionWorkflowRunTokenByRunnerID(ctx, run.Install.RunnerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflow run token")
	}

	return map[string]string{
		"NUON_ORG_ID":     run.OrgID,
		"NUON_APP_ID":     run.Install.AppID,
		"NUON_INSTALL_ID": run.Install.ID,
		"NUON_API_URL":    token.APIURL,
		"NUON_API_TOKEN":  token.Token,
	}, nil
}
