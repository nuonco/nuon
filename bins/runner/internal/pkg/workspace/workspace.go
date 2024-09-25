package workspace

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

const (
	// this is a legacy compatibility value, that was used when we _actually_ didn't need a git repo, but waypoint
	// did not work without having _some_ repo.
	emptyGithubRepoURL string = "https://github.com/jonmorehouse/empty"

	// default tmp root dir to be used when no root is passed in. This allows a user of this workspace to create
	// workspaces in a different directory
	defaultTmpRootDir string = "/tmp"
)

type Workspace interface {
	Init(context.Context) error
	Source() *Source
	Cleanup(context.Context) error
}

type workspace struct {
	v *validator.Validate

	Src        *planv1.GitSource `validate:"required"`
	TmpRootDir string            `validate:"required"`
	ID         string            `validate:"required"`
}

var _ Workspace = (*workspace)(nil)

func New(v *validator.Validate, opts ...workspaceOption) (*workspace, error) {
	obj := &workspace{
		v:          v,
		TmpRootDir: defaultTmpRootDir,
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
func WithGitSource(src *planv1.GitSource) workspaceOption {
	return func(obj *workspace) {
		obj.Src = src
	}
}

// WithWorkspaceID sets an ID on the workspace
func WithWorkspaceID(workspaceID string) workspaceOption {
	return func(obj *workspace) {
		obj.ID = workspaceID
	}
}

// WithTmpRoot sets a root temp directory for the workspace
func WithTmpRoot(root string) workspaceOption {
	return func(obj *workspace) {
		obj.TmpRootDir = root
	}
}
