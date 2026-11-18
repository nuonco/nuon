package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const authorizedOnboardingKey = "nuon.onboarding.authorized"

func (s *service) onboardingTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	onboarding, err := s.getActiveOnboarding(ctx, account.ID)
	if err != nil {
		return permissions.Target{}, err
	}
	if onboarding.OrgID == nil || *onboarding.OrgID != orgID {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	if (onboarding.AppBranchID != nil && *onboarding.AppBranchID != "" || onboarding.InstallID != nil && *onboarding.InstallID != "") &&
		(onboarding.AppID == nil || *onboarding.AppID == "") {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	if onboarding.AppID != nil && *onboarding.AppID != "" {
		var a app.App
		if err := s.db.WithContext(ctx).Where(app.App{ID: *onboarding.AppID, OrgID: orgID}).First(&a).Error; err != nil {
			return permissions.Target{}, err
		}
		if onboarding.AppBranchID != nil && *onboarding.AppBranchID != "" {
			var branch app.AppBranch
			if err := s.db.WithContext(ctx).Where(app.AppBranch{ID: *onboarding.AppBranchID, OrgID: orgID, AppID: a.ID}).First(&branch).Error; err != nil {
				return permissions.Target{}, err
			}
		}
		if onboarding.InstallID != nil && *onboarding.InstallID != "" {
			var install app.Install
			if err := s.db.WithContext(ctx).Where(app.Install{ID: *onboarding.InstallID, OrgID: orgID, AppID: a.ID}).First(&install).Error; err != nil {
				return permissions.Target{}, err
			}
		}
	}
	ctx.Set(authorizedOnboardingKey, onboarding)
	return permissions.Target{ResolvedTarget: permissions.ResolvedTarget{Scope: permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}}}, nil
}
