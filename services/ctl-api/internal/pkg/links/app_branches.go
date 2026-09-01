package links

import (
	"net/url"
)

// AppBranchRunUILink builds the dashboard URL for a single app branch run.
// Callers outside the request path (temporal activities, for one) have no
// context-carried config, so the app URL is passed explicitly.
func AppBranchRunUILink(appURL, orgID, appID, appBranchID, runID string) string {
	if appURL == "" || orgID == "" || appID == "" || appBranchID == "" || runID == "" {
		return ""
	}

	link, err := url.JoinPath(appURL, orgID, "apps", appID, "branches", appBranchID, "runs", runID)
	if err != nil {
		return ""
	}

	return link
}
