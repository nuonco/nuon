package permissions

const (
	ActionInspect         PermissionID = "action.inspect"
	ActionCreate          PermissionID = "action.create"
	ActionUpdate          PermissionID = "action.update"
	ActionDelete          PermissionID = "action.delete"
	ActionConfigCreate    PermissionID = "action.config.create"
	InstallActionInspect  PermissionID = "install.action.inspect"
	ActionAdhocExecute    PermissionID = "action.adhoc.execute"
	ActionRunLogsDelete   PermissionID = "action.run.logs.delete"
	RunbookInspect        PermissionID = "runbook.inspect"
	RunbookCreate         PermissionID = "runbook.create"
	RunbookUpdate         PermissionID = "runbook.update"
	RunbookDelete         PermissionID = "runbook.delete"
	RunbookConfigCreate   PermissionID = "runbook.config.create"
	InstallRunbookInspect PermissionID = "install.runbook.inspect"
	NotebookInspect       PermissionID = "notebook.inspect"
	NotebookCreate        PermissionID = "notebook.create"
	NotebookUpdate        PermissionID = "notebook.update"
	NotebookDelete        PermissionID = "notebook.delete"
	NotebookExecute       PermissionID = "notebook.execute"
)

func ExecutionDefinitions() []PermissionDefinition {
	app := []ResourceKind{KindOrg, KindApp}
	install := []ResourceKind{KindOrg, KindApp, KindInstall}
	definition := []ResourceKind{KindApp, KindOperation}
	execution := []ResourceKind{KindInstall, KindOperation}
	return []PermissionDefinition{
		{ID: ActionInspect, Title: "Inspect action definitions", Description: "Read app action definitions, labels and executable configuration, including embedded records.", Scopes: app, Targets: definition},
		{ID: ActionCreate, Title: "Create action definitions", Description: "Create an action definition in an app.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: ActionUpdate, Title: "Update action definitions", Description: "Update action definition metadata and labels.", Scopes: app, Targets: definition},
		{ID: ActionDelete, Title: "Delete action definitions", Description: "Delete app action definitions and their existing cleanup effects.", Scopes: app, Targets: definition},
		{ID: ActionConfigCreate, Title: "Configure action code", Description: "Author executable action configuration and its automatic triggers for an app.", Scopes: app, Targets: definition},
		{ID: InstallActionInspect, Title: "Inspect install actions", Description: "Read installed actions, runs and steps with their existing embedded configuration, jobs and log metadata.", Scopes: install, Targets: execution},
		{ID: ActionAdhocExecute, Title: "Execute ad-hoc action", Description: "Submit arbitrary action code for execution on an install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: ActionRunLogsDelete, Title: "Delete action run logs", Description: "Delete terminal action-run logs, copied parent records and spans, and revoke the stream writer; legacy access requires org admin.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: RunbookInspect, Title: "Inspect runbook definitions", Description: "Read app runbooks and their executable configuration, steps and inputs.", Scopes: app, Targets: definition},
		{ID: RunbookCreate, Title: "Create runbook definitions", Description: "Create an app runbook definition.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: RunbookUpdate, Title: "Update runbook definitions", Description: "Update app runbook metadata and labels.", Scopes: app, Targets: definition},
		{ID: RunbookDelete, Title: "Delete runbook definitions", Description: "Delete an app runbook subject to existing reference constraints.", Scopes: app, Targets: definition},
		{ID: RunbookConfigCreate, Title: "Configure runbook code", Description: "Author executable runbook steps and automatic triggers for an app.", Scopes: app, Targets: definition},
		{ID: InstallRunbookInspect, Title: "Inspect install runbooks", Description: "Read installed runbooks, rendered configuration and runs with their embedded workflows.", Scopes: install, Targets: execution},
		{ID: NotebookInspect, Title: "Inspect notebooks", Description: "Read install notebooks, executable cells and runs with their execution and log linkage.", Scopes: install, Targets: execution},
		{ID: NotebookCreate, Title: "Create notebooks", Description: "Create an install notebook and start its warm workflow.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: NotebookUpdate, Title: "Update notebooks", Description: "Edit notebook metadata, author executable cells and reorder cells.", Scopes: install, Targets: execution},
		{ID: NotebookDelete, Title: "Delete notebooks", Description: "Delete install notebooks or their cells.", Scopes: install, Targets: execution},
		{ID: NotebookExecute, Title: "Execute notebook cells", Description: "Admit a cell execution on its owning install, including idempotent resubmission; accepted work continues under the existing execution contract.", Scopes: install, Targets: []ResourceKind{KindInstall}},
	}
}
