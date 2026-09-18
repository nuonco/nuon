package vcs

import (
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/vcserrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const referenceNotFoundSignal = "reference not found"

var gitRefContextSignals = []string{
	"unable to clone repo",
	"failed to check out as",
	"unable to check out",
}

var cloneWithRefPattern = regexp.MustCompile(`unable to clone repo\s+(\S+)\s+with ref\s+(\S+)`)

func parseGitRefNotFound(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	if !strings.Contains(ctx.Raw, referenceNotFoundSignal) {
		return nil
	}

	hasContext := false
	for _, signal := range gitRefContextSignals {
		if strings.Contains(ctx.Raw, signal) {
			hasContext = true
			break
		}
	}
	if !hasContext {
		return nil
	}

	var repo, ref string
	if match := cloneWithRefPattern.FindStringSubmatch(ctx.Raw); match != nil {
		repo = normalizeRepo(match[1])
		ref = strings.Trim(match[2], `"'`)
	}

	return &vcserrors.GitRefNotFoundError{Repo: repo, Ref: ref}
}

func init() {
	errparse.Register(errparse.NewParser(errparse.LayerToolSpecific, parseGitRefNotFound,
		errparse.WithSignals(referenceNotFoundSignal),
	))
}
