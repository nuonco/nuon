package permissions

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
)

const (
	OrgInspect             PermissionID = "org.inspect"
	AppInspect             PermissionID = "app.inspect"
	InstallInspect         PermissionID = "install.inspect"
	InstallCreate          PermissionID = "install.create"
	InstallUpdate          PermissionID = "install.update"
	RunnerJobInspect       PermissionID = "runner_job.inspect"
	RunnerJobCancel        PermissionID = "runner_job.cancel"
	LogStreamAppend        PermissionID = "log_stream.append"
	WorkflowInspect        PermissionID = "workflow.inspect"
	WorkflowCancel         PermissionID = "workflow.cancel"
	WorkflowRetry          PermissionID = "workflow.retry"
	WorkflowApprove        PermissionID = "workflow.approve"
	InstallProvision       PermissionID = "install.provision"
	InstallDeploy          PermissionID = "install.deploy"
	InstallTeardown        PermissionID = "install.teardown"
	InstallInputsUpdate    PermissionID = "install.inputs.update"
	InstallConfigUpdate    PermissionID = "install.config.update"
	InstallSecretsSync     PermissionID = "install.secrets.sync"
	InstallStateRead       PermissionID = "install.state.read"
	InstallConfigRead      PermissionID = "install.config.read"
	InstallHistoryRead     PermissionID = "install.history.read"
	InstallDriftResolve    PermissionID = "install.drift.resolve"
	InstallRunnerBootstrap PermissionID = "install.runner.bootstrap"
	InstallCollectorManage PermissionID = "install.collector.manage"
	ActionExecute          PermissionID = "action.execute"
	RunbookExecute         PermissionID = "runbook.execute"
	PolicyReportInspect    PermissionID = "policy_report.inspect"
	PolicyAnalyticsRead    PermissionID = "policy_analytics.read"
	AppBundleRead          PermissionID = "app.bundle.read"
	AppBundleCreate        PermissionID = "app.bundle.create"
	AppBundleDownload      PermissionID = "app.bundle.download"
	QueueInspect           PermissionID = "queue.inspect"
	QueueGraphRead         PermissionID = "queue.graph.read"
)

type PermissionDefinition struct {
	ID          PermissionID   `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Scopes      []ResourceKind `json:"scopes"`
	Targets     []ResourceKind `json:"targets,omitempty"`
	Members     []PermissionID `json:"members,omitempty"`
}

type Catalog struct {
	definitions map[PermissionID]PermissionDefinition
	expanded    map[PermissionID][]PermissionID
	revision    string
}

func DefaultCatalog() (*Catalog, error) {
	org := []ResourceKind{KindOrg}
	app := []ResourceKind{KindOrg, KindApp}
	install := []ResourceKind{KindOrg, KindApp, KindInstall}
	job := []ResourceKind{KindOrg, KindApp, KindInstall, KindRunnerJob}
	stream := []ResourceKind{KindOrg, KindApp, KindInstall, KindRunnerJob, KindLogStream}
	definitions := []PermissionDefinition{
		{ID: OrgInspect, Title: "Inspect org", Description: "Read the org inspection response.", Scopes: org, Targets: []ResourceKind{KindOrg}},
		{ID: AppInspect, Title: "Inspect app", Description: "Read the app inspection response.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: InstallInspect, Title: "Inspect install", Description: "Read the install inspection response, including embedded metadata.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallUpdate, Title: "Update install", Description: "Update an existing install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallDelete, Title: "Delete install", Description: "Delete, deprovision or forget an existing install, including its teardown workflow and install-owned cleanup.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallProvision, Title: "Provision install", Description: "Provision or reprovision an install's infrastructure, stack and sandbox.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallDeploy, Title: "Deploy install", Description: "Deploy or recover components on an install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallTeardown, Title: "Teardown install resources", Description: "Teardown or forget an install's components or sandbox without deleting the install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallInputsUpdate, Title: "Update install inputs", Description: "Change install inputs and run their update workflow.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallConfigUpdate, Title: "Update install configuration", Description: "Change and synchronize install configuration, including its selected app configuration.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallSecretsSync, Title: "Sync install secrets", Description: "Synchronize secrets into an install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallStateRead, Title: "Read install state", Description: "Read current and historical install state, resource status and drift, with existing redaction.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallConfigRead, Title: "Read install configuration", Description: "Read install inputs, rendered readme, cloud roles, stack outputs and CLI configuration with existing redaction.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallHistoryRead, Title: "Read install history", Description: "Read deploys, sandbox and stack runs, events and their existing embedded records.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallCreate, Title: "Create installs", Description: "Create and provision installs for an app, including its initial managed identities.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: InstallDriftResolve, Title: "Resolve install drift", Description: "Reconcile install drift by deploying components, reprovisioning infrastructure or applying inputs, with the existing approved-plan reuse behavior.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallRunnerBootstrap, Title: "Bootstrap install runner", Description: "Issue a bootstrap credential for the install's runner.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallCollectorManage, Title: "Manage install collector", Description: "Create or revoke the install's managed telemetry collector identity and grants.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: ActionExecute, Title: "Execute install action", Description: "Run an action on an install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: RunbookExecute, Title: "Execute install runbook", Description: "Run a runbook on an install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: WorkflowInspect, Title: "Inspect workflow", Description: "Read workflows and their steps, approvals and plan contents.", Scopes: install, Targets: []ResourceKind{KindWorkflow}},
		{ID: WorkflowCancel, Title: "Cancel workflow", Description: "Cancel a workflow or one of its steps.", Scopes: install, Targets: []ResourceKind{KindWorkflow}},
		{ID: WorkflowRetry, Title: "Retry workflow", Description: "Retry or skip a workflow step; also requires authority for the underlying operation.", Scopes: install, Targets: []ResourceKind{KindWorkflow}},
		{ID: WorkflowApprove, Title: "Approve workflow", Description: "Respond to approvals or change approval settings; also requires authority for the underlying operation.", Scopes: install, Targets: []ResourceKind{KindWorkflow}},
		{ID: RunnerJobInspect, Title: "Inspect job", Description: "Read the job inspection response.", Scopes: job, Targets: []ResourceKind{KindRunnerJob}},
		{ID: RunnerJobCancel, Title: "Cancel job", Description: "Cancel an existing job, subject to its state constraints.", Scopes: job, Targets: []ResourceKind{KindRunnerJob}},
		{ID: LogStreamRead, Title: "Read log stream", Description: "Read the whole stream, including shared or copied records.", Scopes: stream, Targets: []ResourceKind{KindLogStream}},
		{ID: LogStreamAppend, Title: "Append log stream", Description: "Append records with the existing parent-copy side effects.", Scopes: []ResourceKind{KindOrg, KindApp, KindInstall, KindLogStream}, Targets: []ResourceKind{KindLogStream}},
		{ID: PolicyReportInspect, Title: "Inspect policy reports", Description: "Read and export reports for verified build, deploy and sandbox-run owners.", Scopes: install, Targets: []ResourceKind{KindOperation}},
		{ID: PolicyAnalyticsRead, Title: "Read policy analytics", Description: "Read aggregate policy results across an app, including its installs; not available at install scope.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppBundleRead, Title: "Read app bundle metadata", Description: "Read publish status and metadata for app bundles.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppBundleCreate, Title: "Create app bundles", Description: "Publish an archive from an app's pinned configuration and build artifacts.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: AppBundleDownload, Title: "Download app bundles", Description: "Issue a presigned download URL for an app archive; issued URLs retain their existing expiry behavior.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: QueueInspect, Title: "Inspect queues", Description: "Read queues, status and signals under their verified owner, including existing embedded signal payloads.", Scopes: install, Targets: []ResourceKind{KindOperation}},
		{ID: QueueGraphRead, Title: "Read execution graphs", Description: "Read recursive execution graphs that can cross queue ownership boundaries within the org.", Scopes: org, Targets: []ResourceKind{KindOrg}},
		{ID: RunnerJobRead, Title: "Job Reader", Description: "Read jobs and their validated associated streams.", Scopes: []ResourceKind{KindRunnerJob}, Members: []PermissionID{RunnerJobInspect, LogStreamRead}},
		{ID: InstallRead, Title: "Install Reader", Description: "Read installs and their owned state, configuration, history, workflows, jobs and logs.", Scopes: []ResourceKind{KindInstall}, Members: []PermissionID{InstallInspect, InstallStateRead, InstallConfigRead, InstallHistoryRead, InstallActionInspect, InstallRunbookInspect, NotebookInspect, WorkflowInspect, RunnerInspect, RunnerJobRead, LogStreamRead, PolicyReportInspect, QueueInspect, TerraformRead}},
		{ID: InstallWrite, Title: "Install write", Description: "Broad install writes, including destructive lifecycle, execution, runner bootstrap and collector identities; not an Editor preset.", Scopes: []ResourceKind{KindInstall}, Members: []PermissionID{InstallUpdate, InstallDelete, InstallProvision, InstallDeploy, InstallTeardown, InstallInputsUpdate, InstallConfigUpdate, InstallSecretsSync, InstallDriftResolve, InstallRunnerBootstrap, InstallCollectorManage, ActionExecute, ActionAdhocExecute, RunbookExecute, NotebookCreate, NotebookUpdate, NotebookDelete, NotebookExecute, WorkflowCancel, WorkflowRetry, WorkflowApprove, RunnerJobCancel, RunnerControl, RunnerSettingsUpdate, RunnerTokensPrune, TerraformCreate, TerraformDelete, TerraformLock, TerraformStateWrite}},
		{ID: AppRead, Title: "App Reader", Description: "Read apps and their supported owned installs, workflows, jobs and logs.", Scopes: []ResourceKind{KindApp}, Members: []PermissionID{AppInspect, AppDefinitionRead, AppBundleRead, ActionInspect, RunbookInspect, InstallRead, WorkflowInspect, RunnerJobRead, LogStreamRead, PolicyReportInspect, PolicyAnalyticsRead}},
		{ID: OrgRead, Title: "Org Reader", Description: "Read the org and supported owned apps, installs, workflows, jobs and logs.", Scopes: org, Members: []PermissionID{OrgInspect, OrgMembersRead, OrgInvitesRead, OrgStatsRead, OrgFeaturesRead, OrgWebhooksRead, RolesRead, AppRead, InstallRead, WorkflowInspect, RunnerJobRead, LogStreamRead, QueueInspect, QueueGraphRead, CloudConnectionRead, ImageRepositoryRead, VCSConnectionRead, SlackRead, TriggerRead, ExecutionConfigRead, StackConfigRead, TelemetrySettingsRead, RunnerComponentHealthRead, RunnerQueueRead, RunnerSandboxConfigRead, HelmReleaseRead}},
	}
	definitions = append(definitions, OrgDefinitions()...)
	definitions = append(definitions, AppDefinitions()...)
	definitions = append(definitions, ExecutionDefinitions()...)
	definitions = append(definitions, IntegrationDefinitions()...)
	definitions = append(definitions, RunnerDefinitions()...)
	definitions = append(definitions, RunnerProtocolDefinitions()...)
	definitions = append(definitions, WorkloadDefinitions()...)
	definitions = append(definitions, AccessDefinitions()...)
	definitions = append(definitions, TriggerDefinitions()...)
	return NewCatalog(definitions)
}

func NewCatalog(definitions []PermissionDefinition) (*Catalog, error) {
	c := &Catalog{
		definitions: make(map[PermissionID]PermissionDefinition, len(definitions)),
		expanded:    make(map[PermissionID][]PermissionID, len(definitions)),
	}
	for _, definition := range definitions {
		if !validScopeID(string(definition.ID)) || definition.Title == "" || definition.Description == "" {
			return nil, fmt.Errorf("permission definition requires a stable ID, title and description")
		}
		if _, exists := c.definitions[definition.ID]; exists {
			return nil, fmt.Errorf("duplicate permission definition %q", definition.ID)
		}
		if len(definition.Scopes) == 0 || (len(definition.Targets) == 0) == (len(definition.Members) == 0) {
			return nil, fmt.Errorf("permission %q requires scopes and either targets or bundle members", definition.ID)
		}
		for _, kind := range append(slices.Clone(definition.Scopes), definition.Targets...) {
			if err := kind.Validate(); err != nil {
				return nil, fmt.Errorf("permission %q: %w", definition.ID, err)
			}
		}
		for _, scope := range definition.Scopes {
			for _, target := range definition.Targets {
				if !scopeSupports(scope, target, definition.ID) {
					return nil, fmt.Errorf("permission %q has no matcher from %q to %q", definition.ID, scope, target)
				}
			}
		}
		definition.Scopes = sortedUnique(definition.Scopes)
		definition.Targets = sortedUnique(definition.Targets)
		definition.Members = sortedUnique(definition.Members)
		c.definitions[definition.ID] = definition
	}
	for id := range c.definitions {
		if _, err := c.expand(id, make(map[PermissionID]bool)); err != nil {
			return nil, err
		}
	}
	data, err := json.Marshal(c.Definitions())
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	c.revision = hex.EncodeToString(hash[:])
	return c, nil
}

func (c *Catalog) expand(id PermissionID, visiting map[PermissionID]bool) ([]PermissionID, error) {
	if expanded, ok := c.expanded[id]; ok {
		return expanded, nil
	}
	definition, ok := c.definitions[id]
	if !ok {
		return nil, fmt.Errorf("unknown permission reference %q", id)
	}
	if visiting[id] {
		return nil, fmt.Errorf("permission bundle cycle at %q", id)
	}
	visiting[id] = true
	operations := []PermissionID{id}
	if len(definition.Members) != 0 {
		operations = nil
		for _, member := range definition.Members {
			leaves, err := c.expand(member, visiting)
			if err != nil {
				return nil, err
			}
			operations = append(operations, leaves...)
		}
		for _, operation := range operations {
			for _, scope := range definition.Scopes {
				if !slices.Contains(c.definitions[operation].Scopes, scope) {
					return nil, fmt.Errorf("bundle %q at %q scope cannot include %q", id, scope, operation)
				}
			}
		}
	}
	delete(visiting, id)
	c.expanded[id] = sortedUnique(operations)
	return c.expanded[id], nil
}

func (c *Catalog) Revision() string {
	return c.revision
}

func (c *Catalog) Definitions() []PermissionDefinition {
	definitions := make([]PermissionDefinition, 0, len(c.definitions))
	for _, definition := range c.definitions {
		definition.Scopes = slices.Clone(definition.Scopes)
		definition.Targets = slices.Clone(definition.Targets)
		definition.Members = slices.Clone(definition.Members)
		definitions = append(definitions, definition)
	}
	slices.SortFunc(definitions, func(a, b PermissionDefinition) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return definitions
}

func sortedUnique[T ~string](values []T) []T {
	values = slices.Clone(values)
	slices.Sort(values)
	return slices.Compact(values)
}
