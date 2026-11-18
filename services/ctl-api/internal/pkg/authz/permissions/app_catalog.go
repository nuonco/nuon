package permissions

const (
	AppCreate             PermissionID = "app.create"
	AppUpdate             PermissionID = "app.update"
	AppAutomationManage   PermissionID = "app.automation.manage"
	AppConfigRead         PermissionID = "app.config.read"
	AppConfigWrite        PermissionID = "app.config.write"
	AppConfigApply        PermissionID = "app.config.apply"
	AppSecretsRead        PermissionID = "app.secrets.read"
	AppSecretsWrite       PermissionID = "app.secrets.write"
	AppBranchRead         PermissionID = "app.branch.read"
	AppBranchWrite        PermissionID = "app.branch.write"
	AppBranchDelete       PermissionID = "app.branch.delete"
	AppBuildRead          PermissionID = "app.build.read"
	AppBuildCreate        PermissionID = "app.build.create"
	AppDefinitionRead     PermissionID = "app.definition.read"
	AppDefinitionWrite    PermissionID = "app.definition.write"
	ComponentInspect      PermissionID = "component.inspect"
	ComponentCreate       PermissionID = "component.create"
	ComponentUpdate       PermissionID = "component.update"
	ComponentDelete       PermissionID = "component.delete"
	ComponentConfigRead   PermissionID = "component.config.read"
	ComponentConfigWrite  PermissionID = "component.config.write"
	AppComponentsBuild    PermissionID = "app.components.build"
	ComponentBuildCreate  PermissionID = "component.build.create"
	ComponentBuildInspect PermissionID = "component_build.inspect"
	ComponentBuildCancel  PermissionID = "component_build.cancel"
)

func AppDefinitions() []PermissionDefinition {
	app := []ResourceKind{KindOrg, KindApp}
	component := []ResourceKind{KindOrg, KindApp, KindComponent}
	build := []ResourceKind{KindOrg, KindApp, KindComponentBuild}
	return []PermissionDefinition{
		{ID: AppCreate, Title: "Create apps", Description: "Create an app in the organization.", Scopes: []ResourceKind{KindOrg}, Targets: []ResourceKind{KindOrg}},
		{ID: AppUpdate, Title: "Update app", Description: "Update app metadata and notification configuration.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppDelete, Title: "Delete apps", Description: "Delete apps and clean up their components and installs under the existing lifecycle constraints. Requires organization-wide authority.", Scopes: []ResourceKind{KindOrg}, Targets: []ResourceKind{KindApp}},
		{ID: AppAutomationManage, Title: "Manage app automation", Description: "Organization-wide configuration activation, connected-repository authoring, builds, branch rollouts, install synchronization and event-driven execution. Authorizes automation's current and future targets within the org; accepted work retains existing execution semantics.", Scopes: []ResourceKind{KindOrg}, Targets: []ResourceKind{KindApp, KindInstall, KindOperation, KindWorkflow}},
		{ID: AppConfigRead, Title: "Read app configuration", Description: "Read app configuration, source archives, graphs and configuration comparisons.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppConfigWrite, Title: "Write app configuration", Description: "Create app-owned configuration records without applying them to installs.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppConfigApply, Title: "Apply app configuration", Description: "Apply configuration to selected installs; each affected install also requires configuration update authority.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppSecretsRead, Title: "Read app secrets", Description: "Read the existing app secret listing response.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppSecretsWrite, Title: "Write app secrets", Description: "Create and delete app secrets.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppBranchRead, Title: "Read app branches", Description: "Read app branches, configuration, previews and run history including embedded install metadata.", Scopes: app, Targets: []ResourceKind{KindApp, KindOperation}},
		{ID: AppBranchWrite, Title: "Write app branch metadata", Description: "Create branches and update their metadata without dispatching branch runs.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppBranchDelete, Title: "Delete app branches", Description: "Delete app branches; deactivating each connected install additionally requires its configuration update authority.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppBuildRead, Title: "Read sandbox builds", Description: "Read app sandbox builds and their embedded metadata.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppBuildCreate, Title: "Create sandbox builds", Description: "Build an app sandbox artifact without applying it to installs.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: ComponentInspect, Title: "Inspect components", Description: "Read component responses including embedded dependencies, configuration and builds.", Scopes: component, Targets: []ResourceKind{KindComponent}},
		{ID: ComponentCreate, Title: "Create components", Description: "Create component definitions in an app.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: ComponentUpdate, Title: "Update components", Description: "Update component metadata and labels.", Scopes: component, Targets: []ResourceKind{KindComponent}},
		{ID: ComponentDelete, Title: "Delete components", Description: "Delete component definitions after existing active-config and install-use constraints permit cleanup.", Scopes: component, Targets: []ResourceKind{KindComponent}},
		{ID: ComponentConfigRead, Title: "Read component configuration", Description: "Read the component's current and historical configuration responses.", Scopes: component, Targets: []ResourceKind{KindComponent}},
		{ID: ComponentConfigWrite, Title: "Write component configuration", Description: "Create component configurations and update component type; automatically dispatched builds additionally require build creation authority.", Scopes: component, Targets: []ResourceKind{KindComponent}},
		{ID: AppComponentsBuild, Title: "Build all app components", Description: "Dispatch builds for the admitted app component set; each component additionally requires build creation authority.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: ComponentBuildCreate, Title: "Create component builds", Description: "Build a component artifact without applying it to installs.", Scopes: component, Targets: []ResourceKind{KindComponent}},
		{ID: ComponentBuildInspect, Title: "Inspect component builds", Description: "Read component build responses including embedded job metadata.", Scopes: build, Targets: []ResourceKind{KindComponentBuild}},
		{ID: ComponentBuildCancel, Title: "Cancel component builds", Description: "Cancel the queue signal of an existing component build.", Scopes: build, Targets: []ResourceKind{KindComponentBuild}},
		{ID: AppDefinitionRead, Title: "Read app definitions", Description: "Read app configuration, branches, secrets, components and builds.", Scopes: app, Members: []PermissionID{AppConfigRead, AppSecretsRead, AppBranchRead, AppBuildRead, ComponentInspect, ComponentConfigRead, ComponentBuildInspect}},
		{ID: AppDefinitionWrite, Title: "Write app definitions", Description: "Update app definitions, including branch/component deletion and builds; applying configuration or disconnecting a branch additionally checks installs.", Scopes: app, Members: []PermissionID{AppUpdate, AppConfigWrite, AppConfigApply, AppSecretsWrite, AppBranchWrite, AppBranchDelete, AppBuildCreate, AppComponentsBuild, ComponentCreate, ComponentUpdate, ComponentDelete, ComponentConfigWrite, ComponentBuildCreate, ComponentBuildCancel}},
	}
}
