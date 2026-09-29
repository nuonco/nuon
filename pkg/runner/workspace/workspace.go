package workspace

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

const (
	emptyGithubRepoURL string = "https://github.com/jonmorehouse/empty"

	DefaultTmpRootDir string = "/tmp"

	HostActionRootDir string = "/opt/nuon/action-workspaces"

	HostActionFallbackRootDir string = "/var/tmp/nuon-action-workspaces"
)

func HostActionRoots() []string {
	return []string{HostActionRootDir, HostActionFallbackRootDir, DefaultTmpRootDir}
}

// why: ResolveHostActionRoot returns the first root the process can create that is
// not memory-backed, so a multi-GiB action writes to disk rather than RAM.
// Landing on a memory-backed root is the failure this exists to prevent, so it
// is reported rather than silently accepted.
func ResolveHostActionRoot(l *zap.Logger) string {
	return resolveActionRoot(l, HostActionRoots())
}

func resolveActionRoot(l *zap.Logger, candidates []string) string {
	var firstUsable string

	for _, root := range candidates {
		if err := os.MkdirAll(root, 0o755); err != nil {
			l.Warn("skipping action workspace root",
				zap.String("path", root),
				zap.Error(err),
			)
			continue
		}

		if firstUsable == "" {
			firstUsable = root
		}

		if fsType := FilesystemType(root); fsType == "tmpfs" || fsType == "ramfs" {
			l.Warn("skipping memory-backed action workspace root",
				zap.String("path", root),
				zap.String("filesystem", fsType),
			)
			continue
		}

		if root != candidates[0] {
			l.Warn("action workspaces are not on the preferred root",
				zap.String("path", root),
				zap.String("preferred", candidates[0]),
			)
		}
		return root
	}

	if firstUsable == "" {
		l.Error("no usable action workspace root, falling back to the default",
			zap.String("path", DefaultTmpRootDir))
		return DefaultTmpRootDir
	}

	l.Error("every action workspace root is memory-backed, so action content will consume RAM instead of disk",
		zap.String("path", firstUsable))

	return firstUsable
}

type Workspace interface {
	Init(context.Context) error
	Source() *Source
	Cleanup(context.Context) error

	Root() string
	AbsPath(string) string
	IsFile(string) bool
	IsDir(string) bool
	RmDir(string) error
	IsExecutable(string) bool
}

type workspace struct {
	v *validator.Validate

	Src *plantypes.GitSource

	TmpRootDir string `validate:"required"`
	ID         string `validate:"required"`

	L *zap.Logger `validate:"required"`
}

var _ Workspace = (*workspace)(nil)

func New(v *validator.Validate, opts ...workspaceOption) (*workspace, error) {
	// TODO(jm): remove this
	l, _ := zap.NewProduction()
	obj := &workspace{
		L:          l,
		v:          v,
		TmpRootDir: DefaultTmpRootDir,
	}

	for _, opt := range opts {
		opt(obj)
	}
	if err := obj.v.Struct(obj); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	return obj, nil
}

type workspaceOption func(*workspace)

func WithGitSource(src *plantypes.GitSource) workspaceOption {
	return func(obj *workspace) {
		obj.Src = src
	}
}

func WithWorkspaceID(workspaceID string) workspaceOption {
	return func(obj *workspace) {
		obj.ID = "workspace-" + workspaceID
	}
}

func WithTmpRoot(root string) workspaceOption {
	return func(obj *workspace) {
		obj.TmpRootDir = root
	}
}

func WithLogger(l *zap.Logger) workspaceOption {
	return func(obj *workspace) {
		obj.L = l
	}
}
