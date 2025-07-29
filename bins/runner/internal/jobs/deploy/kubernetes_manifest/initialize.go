package kubernetes_manifest

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
)

const (
	defaultChartPackageFilename string = "chart.tgz"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	return nil
}
