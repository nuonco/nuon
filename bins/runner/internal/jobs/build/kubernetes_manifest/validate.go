package kubernetes_manifest

import (
	"context"
	"errors"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if h.state.cfg == nil {
		return errors.New("no kubernetes manifest build config")
	}

	switch h.state.cfg.SourceType {
	case "inline":
		if h.state.cfg.InlineManifest == "" {
			return errors.New("inline source type requires manifest content")
		}
	case "kustomize":
		if h.state.cfg.KustomizePath == "" {
			return errors.New("kustomize source type requires kustomize_path")
		}
	default:
		return errors.New("source type must be 'inline' or 'kustomize'")
	}

	return nil
}
