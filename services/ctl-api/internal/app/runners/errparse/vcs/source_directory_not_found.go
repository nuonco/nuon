package vcs

import (
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/vcserrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const walkSignal = "unable to walk"

var (
	walkPathPattern      = regexp.MustCompile(`unable to walk\s+(\S+?):`)
	missingPathPattern   = regexp.MustCompile(`(?:lstat|stat|open)\s+(\S+?):\s*no such file or directory`)
	workspaceRootPattern = regexp.MustCompile(`(?:^|/)workspace-[^/]+/(.*)$`)
)

func parseSourceDirectoryNotFound(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	raw := strings.ReplaceAll(ctx.Raw, `\`, "/")

	walkMatch := walkPathPattern.FindStringSubmatch(raw)
	if walkMatch == nil {
		return nil
	}
	walkPath := strings.TrimSpace(walkMatch[1])

	missingMatch := missingPathPattern.FindStringSubmatch(raw)
	if missingMatch == nil {
		return nil
	}
	if !withinPath(walkPath, strings.TrimSpace(missingMatch[1])) {
		return nil
	}

	relative := workspaceRelativePath(walkPath)
	if relative == "" {
		return nil
	}

	return &vcserrors.SourceDirectoryNotFoundError{Directory: relative}
}

func withinPath(root, path string) bool {
	root = strings.TrimSuffix(root, "/")
	path = strings.TrimSuffix(path, "/")
	return path == root || strings.HasPrefix(path, root+"/")
}

func workspaceRelativePath(path string) string {
	match := workspaceRootPattern.FindStringSubmatch(path)
	if match == nil {
		return ""
	}
	return strings.Trim(match[1], "/")
}

func init() {
	errparse.Register(errparse.NewParser(errparse.LayerToolSpecific, parseSourceDirectoryNotFound,
		errparse.WithSignals(walkSignal),
	))
}
