package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) getBuiltinEnvVars(ctx workflow.Context, run *app.InstallActionWorkflowRun) (map[string]string, error) {
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
		"TRIGGER_TYPE":    string(run.TriggerType),
	}, nil
}

func (p *Planner) getOverrideEnvVars(ctx workflow.Context, run *app.InstallActionWorkflowRun) (map[string]string, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	state, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: run.Install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}
	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	l.Info("rendering environment variables")
	if err := render.RenderMap(&run.RunEnvVars, stateData); err != nil {
		l.Error("error rendering environment vars",
			zap.Any("env-vars", run.RunEnvVars),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render environment variables")
	}

        return generics.ToStringMap(run.RunEnvVars), nil
}
