package notifications

type Type string

const (
	// org notifications
	NotificationsTypeOrgCreated        Type = "org_created"
	NotificationsTypeOrgInvite         Type = "org_invite"
	NotificationsTypeOrgInviteAccepted Type = "org_invite_accepted"

	// app notifications
	NotificationsTypeAppCreated   = "app_created"
	NotificationsTypeFirstAppSync = "first_app_sync"
	NotificationsTypeAppSyncError = "app_sync_error"

	// compponent notifications
	NotificationsTypeComponentBuildFailed = "component_build_failed"

	// install notifications
	NotificationsTypeFirstInstallCreated = "first_install_created"
	NotificationsTypeInstallCreated      = "install_created"

	// install deployment notifications
	NotificationsTypeDeployFailed = "deploy_failed"

	// release notifications
	NotificationsTypeReleaseSucceeded = "release_succeeded"
)

func (n Type) String() string {
	return string(n)
}

func (n Type) EmailTemplateID() string {
	switch n {
	case NotificationsTypeOrgCreated:
		return "clwjuz3gk01o6usztzl5om75i"
	case NotificationsTypeOrgInvite:
		return "clv8uth3t00l710e216ec0qh2"
	case NotificationsTypeOrgInviteAccepted:
		return "clwjv98e8027jdw0gvrdochop"

	default:
		return ""
	}
}

func (n Type) SlackNotificationTemplate() string {
	switch n {
	// org notifications
	case NotificationsTypeOrgCreated:
		return "Org *{{.org_name}}* was created by {{.created_by}}"
	case NotificationsTypeOrgInvite:
		return "{{.created_by}} invited {{.email}} to *{{.org_name}}*"
	case NotificationsTypeOrgInviteAccepted:
		return "{{.email}} accepted invite to {{.org_name}}"

		// app notifications
	case NotificationsTypeAppCreated:
		return "*{{.created_by}}* created a new app *{{.app_name}}*"
	case NotificationsTypeAppSyncError:
		return "{{.created_by}} had a failed config sync for app *{{.app_name}}*"
	case NotificationsTypeFirstAppSync:
		return "{{.created_by}} synced their first config for app *{{.app_name}}*"

		// compponent notifications
	case NotificationsTypeComponentBuildFailed:
		return "Build of component *{{.component_name}} for app *{{.app_name}}* failed (initiated by {{.created_by}})"

		// install notifications
	case NotificationsTypeFirstInstallCreated:
		return "{{.created_by}} created the first install ({{.install_name}}) for *{{.app_name}}*"
	case NotificationsTypeInstallCreated:
		return "{{.created_by}} created a new install of *{{.app_name}}*"

		// install deployment notifications
	case NotificationsTypeDeployFailed:
		return "Deployment of install *{{.install_name}} for app *{{.app_name}}* failed (initiated by {{.created_by}})"

		// release notifications
	case NotificationsTypeReleaseSucceeded:
		return "Release of app *{{.app_name}}* succeeded (initiated by {{.created_by}})"

	default:
		return ""
	}
}
