package workflow

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) prepareInlineContentsCommand(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig) (string, error) {
	if cfg.InlineContents == "" {
		l.Error("no inline contents were declared in action step config")
		return "", errors.New("no command was defined in action step config")
	}

	fp := h.state.workspace.AbsPath(fmt.Sprintf(".inline-contents-step-%d", cfg.Idx))
	if err := os.WriteFile(fp, []byte(cfg.InlineContents), 0o755); err != nil {
		return "", errors.Wrap(err, "unable to write inline contents")
	}

	return fp, nil
}
