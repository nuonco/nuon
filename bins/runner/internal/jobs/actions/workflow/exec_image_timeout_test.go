package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
)

type deadlineRecordingLauncher struct {
	deadline    time.Time
	hasDeadline bool
}

func (l *deadlineRecordingLauncher) Prepare(ctx context.Context, _ launcher.PrepareSpec) error {
	l.deadline, l.hasDeadline = ctx.Deadline()
	return errors.New("stop before the steps run")
}

func (l *deadlineRecordingLauncher) Release(_ string) {}

func (l *deadlineRecordingLauncher) Run(_ context.Context, _ launcher.RunSpec) error {
	return nil
}

func TestExecPullDeadlineIsNotTheActionTimeout(t *testing.T) {
	lnchr := &deadlineRecordingLauncher{}
	h := &handler{
		launcher: lnchr,
		state: &handlerState{
			plan: &plantypes.ActionWorkflowRunPlan{
				Timeout:        time.Second,
				SourceImage:    "ghcr.io/acme/tools:v1",
				ImageDigestRef: "reg/acme/tools@sha256:abc",
				ImageRegistry: &configs.OCIRegistryRepository{
					RegistryType: configs.OCIRegistryTypePublicOCI,
					OCIAuth:      &configs.OCIRegistryAuth{},
				},
			},
			run: &models.AppInstallActionWorkflowRun{},
		},
	}

	ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
	err := h.Exec(ctx, &models.AppRunnerJob{}, &models.AppRunnerJobExecution{ID: "jexec"})

	require.Error(t, err)
	require.True(t, lnchr.hasDeadline)
	assert.Greater(t, time.Until(lnchr.deadline), time.Minute)
}
