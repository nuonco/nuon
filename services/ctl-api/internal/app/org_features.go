package app

type OrgFeature string

const (
	OrgFeatureOrgRunner                OrgFeature = "org-runner"
	OrgFeatureAppBranches              OrgFeature = "app-branches"
	OrgFeatureUserManagedFeatures      OrgFeature = "user-managed-features"
	OrgFeatureSupportRole              OrgFeature = "support-role"
	OrgFeatureInstallRename            OrgFeature = "install-rename"
	OrgFeatureTerraformProviderMirror  OrgFeature = "terraform-provider-mirror"
	OrgFeatureAppBranchesUI            OrgFeature = "app-branches-ui"
	OrgFeatureTraceView                OrgFeature = "trace-view"
	OrgFeatureStateGenV2               OrgFeature = "state-gen-v2"
	OrgFeatureAutoSkipNoop             OrgFeature = "auto-skip-noop"
	OrgFeatureSlack                    OrgFeature = "slack"
	OrgFeaturePulumiSandbox            OrgFeature = "pulumi-sandbox"
	OrgFeaturePulumiUpdatePlans        OrgFeature = "pulumi-update-plans"
	OrgFeatureNotebooks                OrgFeature = "notebooks"
	OrgFeatureVersionsUI               OrgFeature = "enable-versions-ui"
	OrgFeatureSpaceliftInstallStacks   OrgFeature = "spacelift-install-stacks"
	OrgFeatureStackTFProvider          OrgFeature = "stack-tf-provider"
	OrgFeatureAWSAccountConnections    OrgFeature = "aws-account-connections"
	OrgFeatureComponentHealth          OrgFeature = "component-health"
	OrgFeatureServiceAccountsAndTokens OrgFeature = "service-accounts-and-tokens"
	OrgFeaturePhoneHomeAuth            OrgFeature = "phone-home-auth"
	OrgFeatureRunbookStudio            OrgFeature = "runbook-studio"
	OrgFeatureCronNamespaceIsolation   OrgFeature = "cron-namespace-isolation"
	OrgFeatureTriggers                 OrgFeature = "triggers"
	OrgFeatureNewAppIA                 OrgFeature = "new-app-ia"
	OrgFeatureOrgHealthcheckSweeps     OrgFeature = "org-healthcheck-sweeps"
	OrgFeatureAppInstallSyncing        OrgFeature = "app-install-syncing"
	OrgFeatureSandboxOCIArtifacts      OrgFeature = "sandbox-oci-artifacts"
	OrgFeatureImageBackedActions       OrgFeature = "image-backed-actions"
	OrgFeatureSimpleIA                 OrgFeature = "simple-ia"
	OrgFeatureDisableAppsSync          OrgFeature = "disable-apps-sync"
)

type OrgFeatureDef struct {
	Name        OrgFeature
	Default     bool
	Description string
	AdminOnly   bool
}

func featureCatalog() []OrgFeatureDef {
	return []OrgFeatureDef{
		{Name: OrgFeatureOrgRunner, Description: "Enable organization-specific runner functionality for executing deployments"},
		{Name: OrgFeatureAppBranches, Default: true, Description: "Support for multiple application branches allowing parallel development and testing"},
		{Name: OrgFeatureUserManagedFeatures, AdminOnly: true, Description: "Allow organization users to manage feature flags through the public API (admin-only flag)"},
		{Name: OrgFeatureSupportRole, Description: "Enable the support role option when inviting users to the organization"},
		{Name: OrgFeatureInstallRename, Description: "Allow renaming installs from the dashboard edit install modal"},
		{Name: OrgFeatureTerraformProviderMirror, Description: "Vendor terraform providers at build time and ship them inside the OCI artifact so install runners can `terraform init` without reaching registry.terraform.io"},
		{Name: OrgFeatureStateGenV2, Default: true, Description: "Use the new queue-based partial state regeneration system instead of the legacy full-regeneration workflow"},
		{Name: OrgFeatureAppBranchesUI, Default: true, Description: "Enable the app branches UI in the dashboard for managing and switching between app branches"},
		{Name: OrgFeatureTraceView, Description: "Enable the trace view tab on action runs, deploys, and sandbox runs to visualize OTEL spans emitted by the runner"},
		{Name: OrgFeatureAutoSkipNoop, Description: "Automatically skip noop plans without requiring approval, overriding per-component skip_noops settings"},
		{Name: OrgFeatureSlack, Description: "Enable the Slack integration, including the Slack link in the dashboard sidebar and per-org Slack workspace/channel subscriptions"},
		{Name: OrgFeaturePulumiSandbox, Description: "Enable Pulumi-typed app sandboxes (sandbox type=pulumi) in addition to Terraform"},
		{Name: OrgFeaturePulumiUpdatePlans, Description: "Pin Pulumi applies to the approved preview via saved update plans; leave off for stacks using helm (the helm Release resource fails plan validation)"},
		{Name: OrgFeatureNotebooks, Description: "Enable install-scoped Notebooks — a Jupyter-style surface where each cell runs a command on the install's runner via a long-lived, warm per-notebook Temporal workflow, skipping the cold install-workflow step tree for near-real-time adhoc execution."},
		{Name: OrgFeatureVersionsUI, Description: "Enable the install app config versions tab in the dashboard, showing the history of config updates and component diffs for each install."},
		{Name: OrgFeatureSpaceliftInstallStacks, Description: "Surface the Spacelift options (blueprint and administrative stack) on the install stack await step, so customers can provision the Terraform install stack through Spacelift instead of running Terraform locally."},
		{Name: OrgFeatureStackTFProvider, Description: "Show the TF Module tab in the install stack await step: directions for the published nuonco/stack/aws Terraform module, which reads its configuration from the API and authenticates with the stack's API token. Additive — the existing CloudFormation and Terraform directions are unchanged. AWS installs only."},
		{Name: OrgFeatureAWSAccountConnections, AdminOnly: true, Description: "Enable organization-owned cross-account AWS connections with external ID trust verification."},
		{Name: OrgFeatureComponentHealth, Description: "Enable the live component resource explorer: the install runner reports the Kubernetes and cloud resources each component manages with per-resource health, surfaced in the install Resources tab."},
		{Name: OrgFeatureServiceAccountsAndTokens, Description: "Enable the API tokens and service accounts management pages in the dashboard settings navigation."},
		{Name: OrgFeaturePhoneHomeAuth, AdminOnly: true, Description: "Require install phone-home requests to carry an HMAC signature derived from a per-install secret, and require a target cloud account identifier (AWS account ID, GCP project ID, or Azure subscription ID) at install creation. Depends on the phone-home CMK and management-role IAM grants being in place."},
		{Name: OrgFeatureRunbookStudio, Description: "Enable the runbook studio in the dashboard — a literate editor for authoring runbook markdown around executable steps with a live install-state preview."},
		{Name: OrgFeatureCronNamespaceIsolation, Description: "Route the org's runner-healthcheck and install cron queues into dedicated Temporal namespaces + task queues polled by their own workers, isolating cron load from the api task queue."},
		{Name: OrgFeatureTriggers, Description: "Enable triggers and payload-driven rules that start app branch runs or install runbooks."},
		{Name: OrgFeatureNewAppIA, Description: "Enable the branch-centric app information architecture in the dashboard: branches as the app landing page, grouped navigation, and the app source header. Requires app-branches-ui."},
		{Name: OrgFeatureOrgHealthcheckSweeps, Description: "Replace per-runner and per-process healthcheck cron emitters with two per-org sweep emitters that check all runners/processes in paginated batches. Toggle via POST /v1/orgs/{org_id}/migrate-healthcheck-sweeps, which also migrates the emitters."},
		{Name: OrgFeatureAppInstallSyncing, Description: "Enable app install config syncing: point an app at a git repo of per-install configs so pushes to that repo sync every install's config and create missing installs behind an approval step. Gates the install syncs API, the VCS push fan-out, and the dashboard install syncs tab."},
		{Name: OrgFeatureSandboxOCIArtifacts, Description: "Build the app sandbox into an OCI artifact during branch runs and resolve sandbox runs against that artifact instead of cloning the sandbox git source. With it off, sandbox runs always clone git."},
		{Name: OrgFeatureImageBackedActions, Description: "Allow actions to declare a container image their steps run inside. Nuon mirrors the image into the install registry and the mng process runs each step's command, inline_contents, or repo-backed script in the image via the mounted actions-supervisor. VM-based runners only."},
		{Name: OrgFeatureSimpleIA, Description: "Enable the simplified dashboard information architecture."},
		{Name: OrgFeatureDisableAppsSync, Description: "Reject standalone `nuon apps sync` (POST /v1/apps/:app_id/configs/:config_id/sync). Use app branches (`nuon branches sync`) instead."},
	}
}

func FeatureCatalog() []OrgFeatureDef {
	return featureCatalog()
}

func GetFeatures() []OrgFeature {
	defs := featureCatalog()
	out := make([]OrgFeature, 0, len(defs))
	for _, def := range defs {
		out = append(out, def.Name)
	}
	return out
}

func DefaultFeatures() map[OrgFeature]bool {
	defs := featureCatalog()
	out := make(map[OrgFeature]bool, len(defs))
	for _, def := range defs {
		out[def.Name] = def.Default
	}
	return out
}

type OrgFeatureInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Forced      bool   `json:"forced"`
}

func GetFeatureDescriptions() map[OrgFeature]string {
	defs := featureCatalog()
	out := make(map[OrgFeature]string, len(defs))
	for _, def := range defs {
		out[def.Name] = def.Description
	}
	return out
}

func GetFeaturesWithDescriptions() []OrgFeatureInfo {
	forced := ForcedFeatures()
	defs := featureCatalog()
	result := make([]OrgFeatureInfo, 0, len(defs))
	for _, def := range defs {
		result = append(result, OrgFeatureInfo{
			Name:        string(def.Name),
			Description: def.Description,
			Forced:      forced[string(def.Name)],
		})
	}
	return result
}

func GetUserManageableFeatures() []OrgFeature {
	forced := ForcedFeatures()
	defs := featureCatalog()
	manageable := make([]OrgFeature, 0, len(defs))
	for _, def := range defs {
		if def.AdminOnly {
			continue
		}
		if forced[string(def.Name)] {
			continue
		}
		manageable = append(manageable, def.Name)
	}
	return manageable
}
