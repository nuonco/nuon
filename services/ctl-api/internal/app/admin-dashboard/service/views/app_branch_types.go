package views

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type AppBranchView struct {
	AppBranch *app.AppBranch `json:"app_branch"`
	OrgName   string         `json:"org_name"`
	AppName   string         `json:"app_name"`
}
