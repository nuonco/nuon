package provision

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/sender"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
)

type notifier interface {
	sendSuccessNotification(ctx context.Context, bucket string, req *installsv1.ProvisionRequest) error
	sendStartNotification(ctx context.Context, bucket string, req *installsv1.ProvisionRequest) error
	sendErrorNotification(ctx context.Context, bucket string, req *installsv1.ProvisionRequest, errMsg string) error
}

var _ notifier = (*notifierImpl)(nil)

type notifierImpl struct {
	sender sender.NotificationSender
}

const errorNotificationTemplate string = `:rotating_light: _error provisioning sandbox_ :rotating_light:
• *s3-path*: s3://%s/%s
• *nuon-id*: _%s_
• *error*: _%s_
`

func (n *notifierImpl) sendErrorNotification(ctx context.Context, bucket string, req *installsv1.ProvisionRequest, errMsg string) error {
	prefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)

	notif := fmt.Sprintf(errorNotificationTemplate,
		bucket,
		prefix,
		req.InstallId,
		errMsg)

	return n.sender.Send(ctx, notif)
}

const successNotificationTemplate string = `:white_check_mark: _successfully provisioned sandbox_ :white_check_mark:
• *s3-path*: s3://%s/%s
• *nuon-id*: _%s_
• *runner-id*: _%s_
• *org-id*: _%s_
`

func (n *notifierImpl) sendSuccessNotification(ctx context.Context, bucket string, req *installsv1.ProvisionRequest) error {
	s3Prefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)
	notif := fmt.Sprintf(successNotificationTemplate,
		bucket,
		s3Prefix,
		req.InstallId,
		req.InstallId,
		req.OrgId)

	return n.sender.Send(ctx, notif)
}

const startNotificationTemplate = `:package: _started provisioning sandbox_ :package:
• *s3-path*: s3://%s/%s
• *role*: _%s_
• *nuon-id*: _%s_
`

// sendStartNotification sends the start notification via the configured sender
func (n *notifierImpl) sendStartNotification(ctx context.Context, bucket string, req *installsv1.ProvisionRequest) error {
	prefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)

	var roleOrSubscription string
	if req.AwsSettings != nil {
		roleOrSubscription = req.AwsSettings.AwsRoleArn
	}
	if req.AzureSettings != nil {
		roleOrSubscription = req.AzureSettings.SubscriptionId
	}

	msg := fmt.Sprintf(startNotificationTemplate, bucket, prefix,
		roleOrSubscription,
		req.InstallId,
	)

	return n.sender.Send(ctx, msg)
}
