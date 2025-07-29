package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	DefaultQueueTimeout     time.Duration = time.Hour * 24
	DefaultAvailableTimeout time.Duration = time.Second * 30
	DefaultExecutionTimeout time.Duration = time.Minute * 5

	DefaultMaxExecutions int = 1
)

func (s *Helpers) getExecutionTimeout(typ app.RunnerJobType) time.Duration {
	timeouts := map[app.RunnerJobType]time.Duration{
		// build timeouts
		app.RunnerJobTypeDockerBuild:          time.Minute * 60,
		app.RunnerJobTypeContainerImageBuild:  time.Minute * 15,
		app.RunnerJobTypeHelmChartBuild:       time.Minute * 5,
		app.RunnerJobTypeTerraformModuleBuild: time.Minute * 5,

		// sync timeouts
		app.RunnerJobTypeOCISync: time.Minute * 15,

		// deploy timeouts
		app.RunnerJobTypeTerraformDeploy:          time.Minute * 60,
		app.RunnerJobTypeHelmChartDeploy:          time.Minute * 15,
		app.RunnerJobTypeKubrenetesManifestDeploy: time.Minute * 15,
		app.RunnerJobTypeJobDeploy:                time.Minute * 15,

		// terraform timeouts
		app.RunnerJobTypeSandboxTerraform: time.Minute * 60,
		app.RunnerJobTypeRunnerTerraform:  time.Minute * 15,
		app.RunnerJobTypeRunnerHelm:       time.Minute * 5,
	}
	timeout, ok := timeouts[typ]
	if ok {
		return timeout
	}

	return DefaultExecutionTimeout
}

func (s *Helpers) getJob(ctx context.Context, jobID string) (*app.RunnerJob, error) {
	var runnerJob app.RunnerJob

	if res := s.db.WithContext(ctx).First(&runnerJob, "id = ?", jobID); res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return &runnerJob, nil
}
