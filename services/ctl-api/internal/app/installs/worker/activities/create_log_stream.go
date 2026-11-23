package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	defaultLogStreamTokenDuration time.Duration = time.Hour
)

type CreateLogStreamRequest struct {
	SandboxRunID        string
	DeployID            string
	ActionWorkflowRunID string
	StepID              string
	InstallID           string
	ParentLogStreamID   string
}

// @temporal-gen-v2 activity
func (a *Activities) CreateLogStream(ctx context.Context, req CreateLogStreamRequest) (*app.LogStream, error) {
	typ := "install_sandbox_runs"
	id := req.SandboxRunID

	if req.DeployID != "" {
		id = req.DeployID
		typ = "install_deploys"
	}

	if req.ActionWorkflowRunID != "" {
		id = req.ActionWorkflowRunID
		typ = "install_action_workflow_runs"
	}

	if req.StepID != "" && id == "" {
		id = req.StepID
		typ = "install_workflow_steps"
	}

	if req.InstallID != "" && id == "" {
		id = req.InstallID
		typ = "installs"
	}

	ls := app.LogStream{
		OwnerType:         typ,
		OwnerID:           id,
		Open:              true,
		ParentLogStreamID: generics.NewNullString(req.ParentLogStreamID),
	}

	res := a.db.WithContext(ctx).Create(&ls)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create log stream")
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org id from context")
	}

	token, err := a.acctClient.IssueLogStreamToken(ctx, orgID, ls.ID, defaultLogStreamTokenDuration)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	// this token is only available on the temporal response, and is not persisted to the log stream object
	ls.WriteToken = token.Token
	ls.RunnerAPIURL = a.cfg.RunnerAPIURL
	return &ls, nil
}
