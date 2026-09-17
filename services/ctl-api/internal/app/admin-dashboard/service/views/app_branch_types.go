package views

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// AppBranchView is the view data for the app branches list page. app.AppBranch
// marks Org and App as json:"-", so their names are surfaced here.
type AppBranchView struct {
	AppBranch *app.AppBranch `json:"app_branch"`
	OrgName   string         `json:"org_name"`
	AppName   string         `json:"app_name"`
}
