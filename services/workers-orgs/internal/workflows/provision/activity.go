package provision

type Activities struct {
	notifier
}

func NewActivities(sender NotificationSender) *Activities {
	return &Activities{
		notifier: &notifierImpl{
			sender: sender,
		},
	}
}
