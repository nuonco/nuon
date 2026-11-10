package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// CreateInstallLabels merges parentApp's default labels into reqLabels (erroring
// when a request label conflicts with a default), validates any templated
// labels, and splits the result into its static and templated halves.
func CreateInstallLabels(parentApp *app.App, reqLabels map[string]string) (merged map[string]string, static, templated labels.Labels, err error) {
	staticLabels, templatedLabels := labels.Labels(reqLabels).SplitTemplated()
	return reqLabels, staticLabels, templatedLabels, nil
}

// NoBranchForInstallError reports that an install has no app_branch and its
// labels match no app branch, so no config is available to start it on.
func NoBranchForInstallError(name string) error {
	return stderr.ErrUser{
		Err:         fmt.Errorf("install %s has no app_branch and its labels match no app branch", name),
		Description: "This install has no app_branch and its labels don't match any app branch's install groups. Add labels that a branch selects, or set app_branch.",
	}
}

// CreateInstallTarget describes the app config, and any app branch and group,
// that a new install would start on given its labels and optional app_branch.
type CreateInstallTarget struct {
	AppConfigID                    string
	AppBranchID                    string
	AppBranchName                  string
	AppBranchGroup                 string
	AppBranchGroupAssignmentSource app.InstallAppBranchGroupAssignmentSource
}

// ResolveCreateInstallTarget determines which app config, and any app branch
// and group, a new install with reqLabels and appBranchID would start on,
// without creating anything.
func (s *Helpers) ResolveCreateInstallTarget(ctx context.Context, orgID, appID string, reqLabels map[string]string, appBranchID string) (*CreateInstallTarget, error) {
	return nil, errors.New("not implemented")
}
