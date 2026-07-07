package activities

import (
	"strconv"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func RunTypeFromEventType(eventType string) app.AppBranchRunType {
	switch eventType {
	case "pull_request":
		return app.AppBranchRunTypeGitPreview
	case "push":
		return app.AppBranchRunTypeGit
	default:
		return app.AppBranchRunTypeManual
	}
}

func BuildRunLabels(req *TriggerAppBranchRunFromVCSPushRequest) labels.Labels {
	l := labels.Labels{}
	if len(req.PusherEmails) > 0 {
		l["pusher_email"] = req.PusherEmails[0]
	}
	if req.SenderLogin != "" {
		l["sender"] = req.SenderLogin
	}
	if req.HeadSHA != "" {
		l["commit"] = req.HeadSHA
	}
	if req.PRNumber != nil {
		l["pr"] = strconv.Itoa(*req.PRNumber)
	}
	if req.EventType != "" {
		l["event_type"] = req.EventType
	}
	return l
}
