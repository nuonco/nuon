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
	// this is a legacy compatibility value, that was used when we _actually_ didn't need a git repo, but waypoint
	// did not work without having _some_ repo.
	emptyGithubRepoURL string = "https://github.com/jonmorehouse/empty"

	// DefaultTmpRootDir is the root used when no root is passed in. This allows a user of this workspace to create
	// workspaces in a different directory
	DefaultTmpRootDir string = "/tmp"

	// HostActionRootDir is the root the mng process uses for image-backed action
	// workspaces. mng runs natively on the VM host, where /tmp is tmpfs sized at
	// a fraction of RAM, so an action writing multi-GiB content there would
	// consume memory rather than disk. This path is on the root volume.
	HostActionRootDir string = "/opt/nuon/action-workspaces"

	// HostActionFallbackRootDir is used when the process can't create
	// HostActionRootDir, which happens when the mng unit's sandboxing leaves /opt
	// read-only. /var/tmp is on the root volume on a default VM image, unlike
	// /tmp, so it keeps action content on disk.
	HostActionFallbackRootDir string = "/var/tmp/nuon-action-workspaces"
)

// HostActionRoots are the roots an image-backed action workspace can land in,
// most preferred first. Cleanup has to sweep all of them, since which one a job
// used depends on what the process could write at the time.
func HostActionRoots() []string {
	return []string{HostActionRootDir, HostActionFallbackRootDir, DefaultTmpRootDir}
}

// ResolveHostActionRoot returns the first root the process can create that is
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

	// helpers
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

// WithGitSource sets a git source
func WithGitSource(src *plantypes.GitSource) workspaceOption {
	return func(obj *workspace) {
		obj.Src = src
	}
}

// WithWorkspaceID sets an ID on the workspace, prefixed for identification.
func WithWorkspaceID(workspaceID string) workspaceOption {
	return func(obj *workspace) {
		obj.ID = "workspace-" + workspaceID
	}
}

// WithTmpRoot sets a root temp directory for the workspace
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
