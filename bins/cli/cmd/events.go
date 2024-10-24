package cmd

import (
	"github.com/powertoolsdev/mono/pkg/analytics/events"
)

// these events will live in the cli unless we need to start using them in another service
const (

	// apps
	AppsListEvent          events.Event = "apps_list"
	AppsGetEvent           events.Event = "apps_get"
	AppsCurrentEvent       events.Event = "apps_current"
	AppsSandboxConfigEvent events.Event = "apps_sandbox_config"
	AppsConfigEvent        events.Event = "apps_config"
	AppsInputConfigEvent   events.Event = "apps_input_config"
	AppsRunnerConfigEvent  events.Event = "apps_runner_config"
	AppsSelectEvent        events.Event = "apps_select"
	AppsUnsetCurrentEvent  events.Event = "apps_unset_current"
	AppsSyncEvent          events.Event = "apps_sync"
	AppsValidateEvent      events.Event = "apps_validate"
	AppsCreateEvent        events.Event = "apps_create"
	AppsDeleteEvent        events.Event = "apps_delete"
	AppsRenameEvent        events.Event = "apps_rename"

	//builds
	BuildsListEvent   events.Event = "builds_list"
	BuildsGetEvent    events.Event = "builds_get"
	BuildsCreateEvent events.Event = "builds_create"
	BuildsLogsEvent   events.Event = "builds_logs"

	//components
	ComponentsListEvent         events.Event = "components_list"
	ComponentsDeleteEvent       events.Event = "components_delete"
	ComponentsGetEvent          events.Event = "components_get"
	ComponentsLatestConfigEvent events.Event = "components_latest_config"
	ComponentsListConfigsEvent  events.Event = "components_list_configs"

	//installers
	InstallersListEvent events.Event = "installers_list"

	//installs
	InstallsListEvent               events.Event = "installs_list"
	InstallsGetEvent                events.Event = "installs_get"
	InstallsCreateEvent             events.Event = "installs_create"
	InstallsDeleteEvent             events.Event = "installs_delete"
	InstallsComponentsEvent         events.Event = "installs_components"
	InstallsGetDeployEvent          events.Event = "installs_get_deploy"
	InstallsDeployLogsEvent         events.Event = "installs_deploy_logs"
	InstallsListDeploysEvent        events.Event = "installs_list_deploys"
	InstallsSandboxRunsEvent        events.Event = "installs_sandbox_runs"
	InstallsSandboxRunlogsEvent     events.Event = "installs_sandbox_runlogs"
	InstallsCurrentInputsEvent      events.Event = "installs_current_inputs"
	InstallsSelectEvent             events.Event = "installs_select"
	InstallsUnsetCurrentEvent       events.Event = "installs_unset_current"
	InstallsReprovisionEvent        events.Event = "installs_reprovision"
	InstallsDeprovisionEvent        events.Event = "installs_deprovision"
	InstallsTeardownComponentsEvent events.Event = "installs_teardown_components"
	InstallsDeployComponentsEvent   events.Event = "installs_deploy_components"

	//login
	LoginEvent events.Event = "login"

	// orgs
	OrgsListEvent               events.Event = "orgs_list"
	OrgsCurrentEvent            events.Event = "orgs_current"
	OrgsListHealthChecksEvent   events.Event = "orgs_list_health_checks"
	OrgsAPITokenEvent           events.Event = "orgs_api_token"
	OrgsByIdEvent               events.Event = "orgs_by_id"
	OrgsListConnectedReposEvent events.Event = "orgs_list_connected_repos"
	OrgsListVCSConnectionsEvent events.Event = "orgs_list_vcs_connections"
	OrgsConnectedGithubEvent    events.Event = "orgs_connected_github"
	OrgsSelectEvent             events.Event = "orgs_select"
	OrgsUnsetCurrentEvent       events.Event = "orgs_unset_current"
	OrgsCreateEvent             events.Event = "orgs_create"
	OrgsPrintConfigEvent        events.Event = "orgs_print_config"
	OrgsInviteEvent             events.Event = "orgs_invite"
	OrgsListInvteesEvent        events.Event = "orgs_list_invitees"

	//releases
	ReleasesListEvent  events.Event = "releases_list"
	ReleasesGetEvent   events.Event = "releases_get"
	ReleasesStepsEvent events.Event = "releases_steps"
	ReleaseCreateEvent events.Event = "releases_create"

	//secrets
	SecretsListEvent   events.Event = "secrets_list"
	SecredsDeleteEvent events.Event = "secrets_delete"
	SecretsCreateEvent events.Event = "secrets_create"

	//version
	VersionEvent events.Event = "version"
)
