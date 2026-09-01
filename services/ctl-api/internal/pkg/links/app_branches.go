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

func ComponentBuildUILink(appURL, orgID, appID, componentID, buildID string) string {
	if appURL == "" || orgID == "" || appID == "" || componentID == "" || buildID == "" {
		return ""
	}

	link, err := url.JoinPath(appURL, orgID, "apps", appID, "components", componentID, "builds", buildID)
	if err != nil {
		return ""
	}

	return link
}

func SandboxBuildUILink(appURL, orgID, appID, buildID string) string {
	if appURL == "" || orgID == "" || appID == "" || buildID == "" {
		return ""
	}

	link, err := url.JoinPath(appURL, orgID, "apps", appID, "sandbox", "builds", buildID)
	if err != nil {
		return ""
	}

	return link
}

func InstallAppBranchRunsUILink(appURL, orgID, installID string) string {
	if appURL == "" || orgID == "" || installID == "" {
		return ""
	}

	link, err := url.JoinPath(appURL, orgID, "installs", installID, "app-branch-runs")
	if err != nil {
		return ""
	}

	return link
}
