package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// CreateInstallLabels merges parentApp's default labels into reqLabels (erroring
// when a request label conflicts with a default), validates any templated
// labels, and splits the result into its static and templated halves.
func CreateInstallLabels(parentApp *app.App, reqLabels map[string]string) (merged map[string]string, static, templated labels.Labels, err error) {
	if len(parentApp.DefaultLabels) > 0 {
		combined := make(map[string]string, len(reqLabels)+len(parentApp.DefaultLabels))
		for k, v := range reqLabels {
			combined[k] = v
		}
		for key, val := range parentApp.DefaultLabels {
			if cur, ok := combined[key]; ok && cur != val {
				return nil, nil, nil, stderr.ErrUser{
					Err:         fmt.Errorf("label %q is a default label defined in the app config and cannot be overridden", key),
					Description: fmt.Sprintf("label %q is a default label defined in the app config and cannot be overridden", key),
				}
			}
			combined[key] = val
		}
		reqLabels = combined
	}

	staticLabels, templatedLabels := labels.Labels(reqLabels).SplitTemplated()
	for key, tmpl := range templatedLabels {
		if err := render.ValidateTextTemplate(tmpl); err != nil {
			return nil, nil, nil, stderr.ErrUser{
				Err:         fmt.Errorf("label %q template is invalid: %w", key, err),
				Description: fmt.Sprintf("label %q uses the interpolation syntax but is not a valid template", key),
			}
		}
	}

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

// InstallAssignment describes the app config, and any app branch and group,
// that a new install would start on given its labels and optional app_branch.
type InstallAssignment struct {
	AppConfigID                    string
	AppBranchID                    string
	AppBranchName                  string
	AppBranchGroup                 string
	AppBranchGroupAssignmentSource app.InstallAppBranchGroupAssignmentSource
}

// ResolveInstallAssignment determines which app config, and any app branch
// and group, a new install with reqLabels and appBranchID would start on,
// without creating anything.
func (s *Helpers) ResolveInstallAssignment(ctx context.Context, orgID, appID string, reqLabels map[string]string, appBranchID string) (*InstallAssignment, error) {
	parentApp := app.App{}
	if err := s.db.WithContext(ctx).Where(app.App{OrgID: orgID}).First(&parentApp, "id = ?", appID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app: %w", err)
	}

	_, staticLabels, _, err := CreateInstallLabels(&parentApp, reqLabels)
	if err != nil {
		return nil, err
	}

	pin, err := s.resolveCreateInstallPin(ctx, orgID, appID, &CreateInstallParams{AppBranchID: appBranchID, Name: "new install"}, staticLabels)
	if err != nil {
		return nil, err
	}

	assignment := &InstallAssignment{AppConfigID: pin.AppConfig.ID}
	if pin.BranchID == "" {
		return assignment, nil
	}

	var branch app.AppBranch
	if err := s.db.WithContext(ctx).First(&branch, "id = ?", pin.BranchID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app branch: %w", err)
	}
	assignment.AppBranchID = branch.ID
	assignment.AppBranchName = branch.Name

	groups, err := s.appsHelpers.LatestConfigInstallGroups(ctx, pin.BranchID)
	if err != nil {
		return nil, err
	}
	candidate := app.Install{Labeled: labels.Labeled{Labels: staticLabels}}
	group, source, err := appshelpers.ResolveInstallGroupAssignment(groups, &candidate)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, appshelpers.NoMatchingInstallGroupError(&candidate)
	}
	assignment.AppBranchGroup = group.Name
	assignment.AppBranchGroupAssignmentSource = source

	return assignment, nil
}
