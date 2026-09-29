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

func (h *handler) parseCommand(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, src *plantypes.GitSource) (string, []string, error) {
	if cfg.Command == "" && cfg.InlineContents == "" {
		l.Error("no command or inline_contents defined in action step config")
		return "", nil, errors.New("no command or inline_contents defined in action step config")
	}

	dirName := git.Dir(src)
	pieces := strings.Split(cfg.Command, " ")
	if len(pieces) < 1 {
		return "", nil, errors.New("empty command passed to step")
	}

	scriptPath := filepath.Join(dirName, src.Path, pieces[0])

	if strings.HasPrefix(pieces[0], "./") {
		l.Info(fmt.Sprintf("looking for script %s inside of step repo", cfg.Command))
		if !h.state.workspace.IsFile(scriptPath) {
			l.Error(fmt.Sprintf("file %s does not exist", cfg.Command))
			return "", nil, fmt.Errorf("script %s does not exist in cloned repo", cfg.Command)
		}

		if !h.state.workspace.IsExecutable(scriptPath) {
			l.Error(fmt.Sprintf("file exists %s but is not executable", cfg.Command))
			return "", nil, fmt.Errorf("script %s does not exist in cloned repo", cfg.Command)
		}

		if len(pieces) > 1 {
			l.Warn("command configured includes spaces, passing additional arguments as command arguments to root script",
				zap.String("root", pieces[0]),
				zap.Any("args", pieces[1:]),
			)
		}

		return h.state.workspace.AbsPath(scriptPath), pieces[1:], nil
	}

	if h.state.workspace.IsExecutable(scriptPath) {
		l.Info("local path found in step repo, using that")
		return h.state.workspace.AbsPath(scriptPath), pieces[1:], nil
	}

	l.Info(fmt.Sprintf("%s not found in local repo, executing as regular command", pieces[0]))

	// why: you can not look this up in the path here, because a vendor could easily control the image and add
	// something else to the env. (IE: by overriding HOME)
	// Execute through a shell so that environment variable expansion, pipes, and other
	// shell features work as expected.
	return "sh", []string{"-c", cfg.Command}, nil
}
