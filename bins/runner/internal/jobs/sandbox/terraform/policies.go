package terraform

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
)

const (
	policiesDirName    string = "kyverno-policies"
	policiesDirVarName string = "kyverno_policy_dir"
)

func (h *handler) writePolicies(ctx context.Context) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// Write policies inside the workspace so they are cleaned up automatically
	// by workspace.Cleanup(). The workspace root is per-job-execution, so
	// concurrent jobs for the same install don't race.
	policyPath := filepath.Join(h.state.workspace.Root(), policiesDirName)

	// Remove the policy directory if it exists from a previous attempt.
	if err := os.RemoveAll(policyPath); err != nil {
		return errors.Wrap(err, "unable to remove policy directory")
	}

	l.Debug("creating temporary directory to write rendered policies into", zap.String("dir", policyPath))
	if err := os.Mkdir(policyPath, 0o750); err != nil {
		return errors.Wrap(err, "unable to write policies to path")
	}

	for name, contents := range h.state.plan.Policies {
		fp := filepath.Join(policyPath, name)
		if err := os.WriteFile(fp, []byte(contents), 0o644); err != nil {
			return errors.Wrap(err, "unable to write policy file")
		}
	}

	l.Debug("setting kyverno_policy_dir var", zap.String("value", policyPath))
	if h.state.plan.Vars != nil {
		h.state.plan.Vars["kyverno_policy_dir"] = policyPath
	}

	return nil
}
