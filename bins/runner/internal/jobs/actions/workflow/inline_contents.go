package workflow

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/git"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) prepareInlineContentsCommand(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig) (string, error) {
	if cfg.InlineContents == "" {
		l.Error("no inline contents were declared in action step config")
		return "", errors.New("no command was defined in action step config")
	}

	contents, err := stepScriptContents(cfg)
	if err != nil {
		return "", err
	}
	return h.writeStepScript(cfg, contents)
}

func (h *handler) prepareStepScript(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig) (string, error) {
	contents, err := stepScriptContents(cfg)
	if err != nil {
		l.Error("no command or inline_contents defined in action step config")
		return "", err
	}
	return h.writeStepScript(cfg, contents)
}

func stepScriptContents(cfg *models.AppActionWorkflowStepConfig) (string, error) {
	if cfg.InlineContents != "" {
		contents := cfg.InlineContents
		if !strings.HasPrefix(contents, "#!") {
			contents = "#!/bin/sh\n" + contents
		}
		return contents, nil
	}
	if cfg.Command == "" {
		return "", errors.New("no command or inline_contents defined in action step config")
	}
	return "#!/bin/sh\n" + cfg.Command + "\n", nil
}

func (h *handler) writeStepScript(cfg *models.AppActionWorkflowStepConfig, contents string) (string, error) {
	fp := h.state.workspace.AbsPath(fmt.Sprintf(".inline-contents-step-%d", cfg.Idx))
	if err := safeWriteFile(fp, []byte(contents), 0o755); err != nil {
		return "", errors.Wrap(err, "unable to write step script")
	}
	return fp, nil
}

// prepareContainerStep resolves the script, working directory, and extra args
// for an image-backed step. Repo-backed steps clone into the workspace (already
// bind-mounted); a ./script from that clone is executed in the clone root, the
// same as host actions.
func (h *handler) prepareContainerStep(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, src *plantypes.GitSource) (scriptHostPath, workdirHostPath string, scriptArgs []string, err error) {
	workdirHostPath = h.state.workspace.Root()
	if src != nil && src.URL != "" {
		workdirHostPath = h.state.workspace.AbsPath(git.Dir(src))
	}

	if cfg.InlineContents != "" {
		scriptHostPath, err = h.prepareStepScript(ctx, l, cfg)
		return scriptHostPath, workdirHostPath, nil, err
	}

	if src == nil {
		src = &plantypes.GitSource{}
	}

	cmd, args, err := h.parseCommand(ctx, l, cfg, src)
	if err != nil {
		return "", "", nil, err
	}

	if isUnderWorkspace(h.state.workspace.Root(), cmd) {
		return cmd, workdirHostPath, args, nil
	}

	scriptHostPath, err = h.prepareStepScript(ctx, l, cfg)
	return scriptHostPath, workdirHostPath, nil, err
}

func isUnderWorkspace(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
