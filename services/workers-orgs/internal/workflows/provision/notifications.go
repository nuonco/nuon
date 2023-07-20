package provision

import (
	"context"
	"fmt"
)

// NotificationSender defines the interface for sending messages
// There are implementation in the `sender` package but others may be created trivially.
type NotificationSender interface {
	Send(context.Context, string) error
}

type notifier interface {
	sendSuccessNotification(context.Context, SendNotificationRequest) error
	sendStartNotification(context.Context, SendNotificationRequest) error
	sendErrorNotification(context.Context, SendNotificationRequest) error
}

type notifierImpl struct {
	sender NotificationSender
}

var errNoValidSender error = fmt.Errorf("no sender specified")

const startNotifTemplate string = `:package: _started provisioning a new org_ :package:
• *nuon-id*: _%s_
`

type SendNotificationRequest struct {
	ID string `json:"id"`

	Started  bool `json:"started"`
	Erred    bool `json:"erred"`
	Finished bool `json:"finished"`

	// values for err
	ErrStep   string `json:"err_step"`
	ErrString string `json:"error_string"`

	// values for success
	WaypointServerAddress string `json:"waypoint_server_address"`
}

func validateSendNotificationRequest(req SendNotificationRequest) error {
	return nil
}

type SendNotificationResponse struct{}

func (a *Activities) SendNotification(ctx context.Context, req SendNotificationRequest) (SendNotificationResponse, error) {
	resp := SendNotificationResponse{}
	if err := validateSendNotificationRequest(req); err != nil {
		return resp, err
	}

	var err error
	if req.Started {
		err = a.sendStartNotification(ctx, req)
	} else if req.Finished {
		err = a.sendSuccessNotification(ctx, req)
	} else {
		err = a.sendErrorNotification(ctx, req)
	}
	return resp, err
}

// sendStartNotification sends the start notification via the configured sender
func (n *notifierImpl) sendStartNotification(ctx context.Context, req SendNotificationRequest) error {
	if n.sender == nil {
		return errNoValidSender
	}

	msg := fmt.Sprintf(startNotifTemplate, req.ID)
	return n.sender.Send(ctx, msg)
}

const successNotifTemplate string = `:checkered_flag: successfully provisioned org :checkered_flag:
• *nuon-id*: _%s_
• *waypoint-server*: _%s_
`

// sendSuccessNotification sends a success notification via the configured sender
func (n *notifierImpl) sendSuccessNotification(ctx context.Context, req SendNotificationRequest) error {
	if n.sender == nil {
		return errNoValidSender
	}

	msg := fmt.Sprintf(successNotifTemplate, req.ID, req.WaypointServerAddress)
	return n.sender.Send(ctx, msg)
}

const errorNotifTemplate string = `:rotating_light: error occurred provisioning org :rotating_light:
• *nuon-id*: _%s_
• *step*: _%s_
%s
`

// sendErrorNotification sends an error notification via the configured sender
func (n *notifierImpl) sendErrorNotification(ctx context.Context, req SendNotificationRequest) error {
	if n.sender == nil {
		return errNoValidSender
	}

	msg := fmt.Sprintf(errorNotifTemplate, req.ID, req.ErrStep, req.ErrString)
	return n.sender.Send(ctx, msg)
}
