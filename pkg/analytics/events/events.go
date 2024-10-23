package events

type Event string

const (
	AppCreated Event = "app_created"

	InstallCreated Event = "install_created"

	InviteSent     Event = "invite_sent"
	InviteAccepted Event = "invite_accepted"

	OrgCreated Event = "org_created"

	// TODO(ja): Not adding these yet, because I'm not sure
	// if we need them. "provision" doesn't have the same significance
	// in all contexts, and it's value to product is unclear.
	//
	// What we really care about are UX interactions and gitting
	// meaninguful goalpoasts. Tracking failed an successful operations
	// is useful for monitoring, but probably not for product.

	// OrgProvisionFailed    Event = "org_provision_failed"
	// OrgProvisionSucceeded Event = "org_provision_succeeded"

	RunnerCreated Event = "runner_created"

	HealthCheckCreated Event = "health_check_created"

	HeartBeatCreated Event = "heart_beat_created"

	CliCommand Event = "cli_command"
)
