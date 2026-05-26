package agent

type ToolDef struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  ToolDefParameters `json:"parameters"`
}

type ToolDefParameters struct {
	Type       string                     `json:"type"`
	Properties map[string]ToolDefProperty `json:"properties,omitempty"`
	Required   []string                   `json:"required,omitempty"`
}

type ToolDefProperty struct {
	Type        string                     `json:"type"`
	Description string                     `json:"description,omitempty"`
	Enum        []string                   `json:"enum,omitempty"`
	Properties  map[string]ToolDefProperty `json:"properties,omitempty"`
	Items       *ToolDefProperty           `json:"items,omitempty"`
}

var AppTools = []ToolDef{
	{
		Name:        "list_apps",
		Description: "List all apps in the current organization.",
		Parameters:  ToolDefParameters{Type: "object"},
	},
	{
		Name:        "get_app",
		Description: "Get details for a specific app.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id": {Type: "string", Description: "The app ID."},
			},
			Required: []string{"app_id"},
		},
	},
	{
		Name:        "create_app",
		Description: "Create a new app. Returns the created app object.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"name": {Type: "string", Description: "Name for the app (e.g. 'my-saas-platform')."},
			},
			Required: []string{"name"},
		},
	},
	{
		Name:        "get_app_config",
		Description: "Get the latest app config with full input definitions including names, types, defaults, and required flags. Use this before creating an install to understand what inputs are needed.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id": {Type: "string", Description: "The app ID."},
			},
			Required: []string{"app_id"},
		},
	},
	{
		Name:        "list_components",
		Description: "List all components in an app.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id": {Type: "string", Description: "The app ID."},
			},
			Required: []string{"app_id"},
		},
	},
	{
		Name:        "create_component",
		Description: "Create a new component in an app.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id": {Type: "string", Description: "The app ID."},
				"name":   {Type: "string", Description: "Name for the component."},
				"kind": {Type: "string", Description: "Component kind.", Enum: []string{
					"terraform_module", "helm_chart", "kubernetes_manifest", "docker_build",
				}},
				"dependencies": {Type: "array", Description: "List of component IDs this depends on.", Items: &ToolDefProperty{Type: "string"}},
			},
			Required: []string{"app_id", "name", "kind"},
		},
	},
	{
		Name:        "create_terraform_config",
		Description: "Create a Terraform module config for a component.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id":       {Type: "string", Description: "The app ID."},
				"component_id": {Type: "string", Description: "The component ID."},
				"connected_repo": {Type: "object", Description: "Git repo connection.", Properties: map[string]ToolDefProperty{
					"repo":      {Type: "string", Description: "Repository full name (org/repo)."},
					"branch":    {Type: "string", Description: "Branch name."},
					"directory": {Type: "string", Description: "Directory path within the repo."},
				}},
				"variables": {Type: "object", Description: "Terraform variable values as key-value pairs."},
				"env_vars":  {Type: "object", Description: "Environment variables as key-value pairs."},
			},
			Required: []string{"app_id", "component_id"},
		},
	},
	{
		Name:        "create_helm_config",
		Description: "Create a Helm chart config for a component.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id":       {Type: "string", Description: "The app ID."},
				"component_id": {Type: "string", Description: "The component ID."},
				"connected_repo": {Type: "object", Description: "Git repo connection.", Properties: map[string]ToolDefProperty{
					"repo":      {Type: "string", Description: "Repository full name (org/repo)."},
					"branch":    {Type: "string", Description: "Branch name."},
					"directory": {Type: "string", Description: "Directory path within the repo."},
				}},
				"values": {Type: "object", Description: "Helm values as key-value pairs."},
			},
			Required: []string{"app_id", "component_id"},
		},
	},
	{
		Name:        "create_k8s_manifest_config",
		Description: "Create a Kubernetes manifest config for a component.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id":       {Type: "string", Description: "The app ID."},
				"component_id": {Type: "string", Description: "The component ID."},
				"connected_repo": {Type: "object", Description: "Git repo connection.", Properties: map[string]ToolDefProperty{
					"repo":      {Type: "string", Description: "Repository full name (org/repo)."},
					"branch":    {Type: "string", Description: "Branch name."},
					"directory": {Type: "string", Description: "Directory path within the repo."},
				}},
				"manifest_contents": {Type: "string", Description: "Inline Kubernetes manifest YAML (alternative to connected_repo)."},
			},
			Required: []string{"app_id", "component_id"},
		},
	},
	{
		Name:        "create_docker_build_config",
		Description: "Create a Docker build config for a component.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id":       {Type: "string", Description: "The app ID."},
				"component_id": {Type: "string", Description: "The component ID."},
				"connected_repo": {Type: "object", Description: "Git repo connection.", Properties: map[string]ToolDefProperty{
					"repo":      {Type: "string", Description: "Repository full name (org/repo)."},
					"branch":    {Type: "string", Description: "Branch name."},
					"directory": {Type: "string", Description: "Directory path within the repo."},
				}},
				"dockerfile": {Type: "string", Description: "Path to Dockerfile within the repo."},
			},
			Required: []string{"app_id", "component_id"},
		},
	},
	{
		Name:        "build_all_components",
		Description: "Trigger builds for all components in an app.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id": {Type: "string", Description: "The app ID."},
			},
			Required: []string{"app_id"},
		},
	},
	{
		Name:        "get_build",
		Description: "Get the status and details of a specific build.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id":       {Type: "string", Description: "The app ID."},
				"component_id": {Type: "string", Description: "The component ID."},
				"build_id":     {Type: "string", Description: "The build ID."},
			},
			Required: []string{"app_id", "component_id", "build_id"},
		},
	},
	{
		Name:        "list_vcs_connections",
		Description: "List VCS (GitHub) connections for the organization.",
		Parameters:  ToolDefParameters{Type: "object"},
	},
	{
		Name:        "list_repos",
		Description: "List repositories available through a VCS connection.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"connection_id": {Type: "string", Description: "The VCS connection ID."},
			},
			Required: []string{"connection_id"},
		},
	},
	{
		Name:        "list_branches",
		Description: "List branches for a repository through a VCS connection.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"connection_id": {Type: "string", Description: "The VCS connection ID."},
				"repo":          {Type: "string", Description: "Repository full name (org/repo)."},
			},
			Required: []string{"connection_id", "repo"},
		},
	},
}

var InstallTools = []ToolDef{
	{
		Name:        "list_installs",
		Description: "List installs in the current organization. Returns up to 100 installs. Use the q parameter to search by name.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"q": {Type: "string", Description: "Optional search query to filter installs by name."},
			},
		},
	},
	{
		Name:        "get_install",
		Description: "Get details for a specific install.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"install_id": {Type: "string", Description: "The install ID."},
			},
			Required: []string{"install_id"},
		},
	},
	{
		Name:        "create_install",
		Description: "Create a new install for an app. Before calling this, use get_app and get_config_template to understand the app's platform, required inputs, and defaults.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"app_id":       {Type: "string", Description: "The app ID to install."},
				"name":         {Type: "string", Description: "Unique name for this install."},
				"region":       {Type: "string", Description: "Cloud region (e.g. us-west-2 for AWS, eastus for Azure)."},
				"location":     {Type: "string", Description: "Azure location (use instead of region for Azure apps)."},
				"auto_approve": {Type: "boolean", Description: "If true, auto-approve all deployments without manual confirmation."},
				"inputs":       {Type: "object", Description: "Input values as key-value string pairs. Use defaults from get_config_template where available; only include inputs that need values."},
			},
			Required: []string{"app_id", "name"},
		},
	},
	{
		Name:        "get_install_inputs",
		Description: "Get the current input values for an existing install. Only use this for inspecting or debugging an existing install — NEVER use this when creating a new install.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"install_id": {Type: "string", Description: "The install ID."},
			},
			Required: []string{"install_id"},
		},
	},
	{
		Name:        "get_cloud_regions",
		Description: "Get available cloud regions for a platform (e.g. AWS, Azure, GCP).",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"platform": {Type: "string", Description: "Cloud platform.", Enum: []string{"aws", "azure", "gcp"}},
			},
			Required: []string{"platform"},
		},
	},
	{
		Name:        "list_deploys",
		Description: "List deploys for an install.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"install_id": {Type: "string", Description: "The install ID."},
			},
			Required: []string{"install_id"},
		},
	},
	{
		Name:        "get_workflows",
		Description: "Get active workflows for an install.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"install_id": {Type: "string", Description: "The install ID."},
			},
			Required: []string{"install_id"},
		},
	},
	{
		Name:        "get_workflow_steps",
		Description: "Get detailed steps for a workflow.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"workflow_id": {Type: "string", Description: "The install workflow ID."},
			},
			Required: []string{"workflow_id"},
		},
	},
	{
		Name:        "get_step_logs",
		Description: "Read logs from a log stream (useful for diagnosing build/deploy failures).",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"log_stream_id": {Type: "string", Description: "The log stream ID."},
			},
			Required: []string{"log_stream_id"},
		},
	},
	{
		Name:        "get_runner_job_plan",
		Description: "Get the Terraform plan output for a runner job.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"runner_job_id": {Type: "string", Description: "The runner job ID."},
			},
			Required: []string{"runner_job_id"},
		},
	},
	{
		Name:        "get_runner",
		Description: "Get details about a runner (the execution agent in a customer's cloud).",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"runner_id": {Type: "string", Description: "The runner ID."},
			},
			Required: []string{"runner_id"},
		},
	},
	{
		Name:        "get_runner_health",
		Description: "Get recent health checks for a runner.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"runner_id": {Type: "string", Description: "The runner ID."},
			},
			Required: []string{"runner_id"},
		},
	},
	{
		Name:        "run_adhoc_action",
		Description: "Run a bash script on an install's runner. The script runs with kubeconfig available so kubectl works. This tool blocks until the action completes and returns the output logs. CRITICAL: Always show the script to the user and get confirmation before calling this tool.",
		Parameters: ToolDefParameters{
			Type: "object",
			Properties: map[string]ToolDefProperty{
				"install_id":      {Type: "string", Description: "The install ID."},
				"inline_contents": {Type: "string", Description: "The bash script to run."},
				"name":            {Type: "string", Description: "Optional label for the action run."},
				"timeout":         {Type: "number", Description: "Timeout in seconds (default 300)."},
			},
			Required: []string{"install_id", "inline_contents"},
		},
	},
}

func AllTools() []ToolDef {
	tools := make([]ToolDef, 0, len(AppTools)+len(InstallTools))
	tools = append(tools, AppTools...)
	tools = append(tools, InstallTools...)
	return tools
}
