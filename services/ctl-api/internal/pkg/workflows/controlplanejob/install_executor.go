package controlplanejob

import (
	"context"
	"fmt"

	runnercontrolplane "github.com/nuonco/nuon/pkg/runner/controlplane"
	"github.com/nuonco/nuon/pkg/runner/launcher"
	launcherkubernetes "github.com/nuonco/nuon/pkg/runner/launcher/kubernetes"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// newExecutor picks the executor for job's runner: the plain build executor
// for every job today, except install jobs whose runner has moved to the
// control plane (RunnerStatusNotUsed), which get the install executor so the
// sandbox and deploy Terraform handlers are available too.
//
// a.cfg.RunnerAPIURL is ctl-api's runner API endpoint -- the same URL a real
// runner would use. Locally that's whatever runner_api_url is configured to
// (internal.Config.RunnerAPIURL). An internal/in-cluster URL would avoid a
// round trip through the public endpoint in a real deployment, but using the
// configured URL keeps this consistent with how a runner always reaches the
// runner API.
func (a *Activities) newExecutor(ctx context.Context, job *app.RunnerJob) (*runnercontrolplane.Executor, error) {
	if job.RunnerID == "" {
		return runnercontrolplane.NewExecutor(a, a.l, runnercontrolplane.Config{GitRef: a.cfg.GitRef})
	}

	var runner app.Runner
	if err := a.db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(&app.Runner{ID: job.RunnerID}).First(&runner).Error; err != nil {
		return nil, fmt.Errorf("unable to get runner job's runner: %w", err)
	}
	if runner.Status != app.RunnerStatusNotUsed {
		return runnercontrolplane.NewExecutor(a, a.l, runnercontrolplane.Config{GitRef: a.cfg.GitRef})
	}

	// Mint the state-backend token right before the run, not ahead of time: a
	// not-used runner holds no token-holding processes, so a token minted
	// earlier would be swept as unheld once it's old enough
	// (minRevokableTokenAge). See IMPLEMENTATION.md item 5.
	token, err := a.runnerHelpers.CreateToken(ctx, job.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to mint install runner token: %w", err)
	}

	actionsLauncher, err := a.actionsLauncher()
	if err != nil {
		return nil, fmt.Errorf("unable to build actions launcher: %w", err)
	}

	return runnercontrolplane.NewInstallExecutor(a, a.l, runnercontrolplane.InstallConfig{
		Config:       runnercontrolplane.Config{GitRef: a.cfg.GitRef},
		RunnerAPIURL: a.cfg.RunnerAPIURL,
		RunnerID:     job.RunnerID,
		APIToken:     token.Token,
		Launcher:     actionsLauncher,
	})
}

// actionsLauncher builds the launcher.Launcher backend for image-backed
// Action steps on runnerless installs, selected by
// RunnerActionsLauncherBackend ("kubernetes", the default for deployed
// control planes on EKS/GKE, or "docker" for local dev on a Mac with Docker).
func (a *Activities) actionsLauncher() (launcher.Launcher, error) {
	switch a.cfg.RunnerActionsLauncherBackend {
	case "docker":
		// local dev only -- see launcher.DockerLauncher's doc comment.
		return launcher.NewDockerLauncher(launcher.NewImageCache()), nil
	case "kubernetes", "":
		return launcherkubernetes.New(launcherkubernetes.Params{
			Namespace: a.cfg.RunnerActionsLauncherNamespace,
			L:         a.l,
		})
	default:
		return nil, fmt.Errorf("unknown runner actions launcher backend %q", a.cfg.RunnerActionsLauncherBackend)
	}
}
