package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type createInstallPin struct {
	AppConfig *app.AppConfig
	BranchID  string
}

// resolveCreateInstallPin decides which app config a new install starts on.
// Omitting a branch always uses the latest unbranched config from apps sync,
// even when the app also has branches.
func (s *Helpers) resolveCreateInstallPin(ctx context.Context, orgID, appID string, req *CreateInstallParams) (*createInstallPin, error) {
	if req.AppBranchID == "" {
		cfg, err := s.latestActiveAppConfig(ctx, orgID, appID, "")
		if err != nil {
			return nil, err
		}
		return &createInstallPin{AppConfig: cfg}, nil
	}

	var branch app.AppBranch
	if err := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: orgID, AppID: appID}).
		First(&branch, "id = ?", req.AppBranchID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("app branch %s not found on app %s", req.AppBranchID, appID),
				Description: "The selected app branch does not belong to this app.",
			}
		}
		return nil, fmt.Errorf("unable to get app branch: %w", err)
	}

	cfg, err := s.latestActiveAppConfig(ctx, orgID, appID, branch.ID)
	if err != nil {
		return nil, err
	}

	return &createInstallPin{
		AppConfig: cfg,
		BranchID:  branch.ID,
	}, nil
}

// LatestActiveBranchAppConfig returns the app config a branch currently
// deploys: its newest active config that is not a pull request preview and
// whose builds are locked. An install joining the branch starts here.
func (s *Helpers) LatestActiveBranchAppConfig(ctx context.Context, appID, branchID string) (*app.AppConfig, error) {
	return s.latestActiveAppConfig(ctx, "", appID, branchID)
}

func (s *Helpers) latestActiveAppConfig(ctx context.Context, orgID, appID, branchID string) (*app.AppConfig, error) {
	query := s.db.WithContext(ctx).
		Preload("SandboxConfig").
		Preload("RunnerConfig").
		Preload("PermissionsConfig").
		Preload("PermissionsConfig.Roles").
		Where(app.AppConfig{AppID: appID}).
		Where(views.TableOrViewName(s.db, &app.AppConfig{}, ".status_v2 ->> 'status' = ?"), string(app.AppConfigStatusActive)).
		Where("labels->>'source' IS NULL OR labels->>'source' != ?", string(app.AppBranchRunTypeGitPreview)).
		Order(views.TableOrViewName(s.db, &app.AppConfig{}, ".created_at DESC"))

	if orgID != "" {
		query = query.Where(app.AppConfig{OrgID: orgID})
	}
	if branchID == "" {
		query = query.Where("app_branch_id IS NULL OR app_branch_id = ''")
	} else {
		query = query.Where(app.AppConfig{AppBranchID: pkggenerics.NewNullString(branchID)})
	}

	deployable := query.Session(&gorm.Session{}).Where(
		"("+views.TableOrViewName(s.db, &app.AppConfig{}, ".builds_pinned = ?")+" OR "+views.TableOrViewName(s.db, &app.AppConfig{}, ".builds_locked = ?")+")",
		false, true,
	)

	var cfg app.AppConfig
	if err := deployable.First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if branchID == "" {
				return nil, stderr.ErrUser{
					Err:         fmt.Errorf("no active app config found for app %s", appID),
					Description: "No active app config found. Please sync your app configuration before creating an install.",
				}
			}
			var building app.AppConfig
			if err := query.Session(&gorm.Session{}).First(&building).Error; err == nil {
				return nil, stderr.ErrUser{
					Err:         fmt.Errorf("app config %s for app branch %s has not finished its builds", building.ID, branchID),
					Description: "The selected app branch is still building. Retry once its branch run completes.",
				}
			}
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("no active app config found for app branch %s", branchID),
				Description: "The selected app branch has no active app config. Sync the branch before creating an install.",
			}
		}
		return nil, fmt.Errorf("unable to get app config: %w", err)
	}

	return &cfg, nil
}
