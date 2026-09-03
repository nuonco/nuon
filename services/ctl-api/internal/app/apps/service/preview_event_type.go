package service

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

func previewEventType(source app.AppBranchRunPreviewSource) string {
	if source == app.AppBranchRunPreviewSourcePR {
		return "pull_request"
	}
	return "manual"
}
