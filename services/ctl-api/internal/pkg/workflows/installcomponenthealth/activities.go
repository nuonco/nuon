package installcomponenthealth

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/runner/componenthealth"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnectionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/helpers"
	runnerhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
)

type Activities struct {
	db               *gorm.DB
	cfg              *internal.Config
	l                *zap.Logger
	cloudConnections cloudConnectionsCredentialsHelper
	runnerHelpers    *runnerhelpers.Helpers
}

// cloudConnectionsCredentialsHelper mirrors
// controlplanejob.cloudConnectionsCredentialsHelper so this package can pass
// *cloudconnectionshelpers.Helpers straight through to
// controlplanejob.InjectAWSAuthCredentials without importing that unexported
// type here.
type cloudConnectionsCredentialsHelper interface {
	Credentials(ctx context.Context, connection *app.CloudConnection, sessionName string) (*awscredentials.Config, error)
}

type ActivityParams struct {
	fx.In

	DB  *gorm.DB `name:"psql"`
	Cfg *internal.Config
	L   *zap.Logger
	// CloudConnections is the concrete fx-provided type (fxmodules.helpers
	// provides cloudconnectionshelpers.New, not the interface below); it is
	// assigned into Activities.cloudConnections, typed as the narrower
	// interface, by ordinary Go interface satisfaction.
	CloudConnections *cloudconnectionshelpers.Helpers
	RunnerHelpers    *runnerhelpers.Helpers
}

func NewActivities(params ActivityParams) *Activities {
	return &Activities{
		db:               params.DB,
		cfg:              params.Cfg,
		l:                params.L,
		cloudConnections: params.CloudConnections,
		runnerHelpers:    params.RunnerHelpers,
	}
}

var _ cloudConnectionsCredentialsHelper = (*cloudconnectionshelpers.Helpers)(nil)

// AllActivities is what Agent E (wave 6) registers on the
// control-plane-installs worker; see registration.go for the full list this
// package needs wired in, including the Workflows side.
func (a *Activities) AllActivities() []any {
	return []any{
		a.InstallComponentHealthReportComponentHealth,
	}
}

// ReportComponentHealthInput is the reportComponentHealth activity's real
// argument. It is not named ReportComponentHealthRequest: temporal-gen-v2
// generates a wrapper type by that exact name (see activities_gen.go), so
// reusing it here for the inner type would collide with the generated one.
type ReportComponentHealthInput struct {
	InstallID string `json:"install_id" validate:"required"`
	RunnerID  string `json:"runner_id" validate:"required"`
}

// reportComponentHealth builds a throwaway componenthealth.Engine scoped to
// one install -- a fresh runner-API token minted per call, the install's
// persisted cluster access rehydrated through the exact same
// GetComponentHealthContext/PutComponentHealthContext round trip a Runner
// uses -- and runs exactly one report cycle through the same
// CreateComponentHealth path. No ticker, no background goroutine: the
// InstallComponentHealthWorkflow supplies the schedule.
//
// A missing runner, or one that has moved off RunnerStatusNotUsed since this
// cycle was scheduled, is not an error: the workflow's stop signal should
// catch up shortly, and skipping a cycle is cheap. Health probes dial
// endpoints directly; see registration.go for which probe types are expected
// to fail for install-private endpoints from the control plane.
//
// @temporal-gen-v2 activity
// @as-wrapper
// @wrapper-prefix InstallComponentHealth
func (a *Activities) reportComponentHealth(ctx context.Context, req *ReportComponentHealthInput) error {
	activity.RecordHeartbeat(ctx, "reporting")

	var runner app.Runner
	if err := a.db.WithContext(ctx).Where(&app.Runner{ID: req.RunnerID}).First(&runner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get runner: %w", err)
	}
	if runner.Status != app.RunnerStatusNotUsed {
		return nil
	}

	token, err := a.runnerHelpers.CreateToken(ctx, req.RunnerID)
	if err != nil {
		return fmt.Errorf("unable to mint install runner token: %w", err)
	}

	installClient, err := nuonrunner.New(
		nuonrunner.WithURL(a.cfg.RunnerAPIURL),
		nuonrunner.WithRunnerID(req.RunnerID),
		nuonrunner.WithAuthToken(token.Token),
	)
	if err != nil {
		return fmt.Errorf("unable to build install runner API client: %w", err)
	}

	l := a.l.With(zap.String("install_id", req.InstallID), zap.String("runner_id", req.RunnerID))

	clusterProvider := componenthealth.NewClusterProvider(componenthealth.ProviderParams{APIClient: installClient, L: l})
	manifestKinds := componenthealth.NewManifestKindsProvider(componenthealth.ManifestKindsProviderParams{L: l, Cluster: clusterProvider})
	terraformProvider := componenthealth.NewTerraformProvider(componenthealth.TerraformProviderParams{L: l, Kinds: manifestKinds})

	// Rehydrate cluster access + discovered kinds persisted by earlier
	// deploys -- the same load a Runner does on boot (Engine.run).
	clusterProvider.Load(ctx)
	manifestKinds.Load()

	if ci := clusterProvider.Get(); ci != nil && ci.AWSAuth != nil {
		// The persisted AssumeRole chain was built for a Runner's own
		// instance identity in the customer's account; the control plane has
		// no such identity, so its two-step source has to be the install's
		// Cloud Connection session instead, exactly like a control-plane
		// deploy/sandbox job's plan (controlplanejob.injectInstallCredentials).
		if err := controlplanejob.InjectAWSAuthCredentials(ctx, a.db, a.cloudConnections, req.InstallID, []*awscredentials.Config{ci.AWSAuth}); err != nil {
			l.Warn("unable to inject install credentials for component health cluster access", zap.Error(err))
		}
	}

	engine, err := componenthealth.New(componenthealth.Params{
		APIClient:     installClient,
		L:             l,
		LC:            noopLifecycle{},
		Process:       "install",
		Cluster:       clusterProvider,
		Terraform:     terraformProvider,
		ManifestKinds: manifestKinds,
	})
	if err != nil {
		return fmt.Errorf("unable to build component health engine: %w", err)
	}

	if err := engine.ReportOnce(ctx); err != nil {
		return fmt.Errorf("unable to report component health: %w", err)
	}
	return nil
}

// noopLifecycle discards Engine's fx.Lifecycle hook instead of running it:
// componenthealth.New only starts its background ticker goroutine from an
// OnStart hook, so never firing one keeps this call to one report cycle, with
// no leaked goroutine.
type noopLifecycle struct{}

func (noopLifecycle) Append(fx.Hook) {}
