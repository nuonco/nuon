package plandiff

// PlanType represents the type of deployment plan
type PlanType string

const (
	PlanTypeTerraform  PlanType = "terraform_plan"
	PlanTypeHelm       PlanType = "helm_approval"
	PlanTypeKubernetes PlanType = "kubernetes_manifest_approval"
	PlanTypeUnknown    PlanType = "unknown"
)

// TerraformChangeAction represents Terraform change actions
type TerraformChangeAction string

const (
	TerraformActionCreate  TerraformChangeAction = "create"
	TerraformActionUpdate  TerraformChangeAction = "update"
	TerraformActionDelete  TerraformChangeAction = "delete"
	TerraformActionNoOp    TerraformChangeAction = "no-op"
	TerraformActionReplace TerraformChangeAction = "replace"
	TerraformActionRead    TerraformChangeAction = "read"
)

// HelmK8sChangeAction represents Helm and Kubernetes change actions
type HelmK8sChangeAction string

const (
	HelmK8sActionAdd       HelmK8sChangeAction = "add"
	HelmK8sActionAdded     HelmK8sChangeAction = "added"
	HelmK8sActionChange    HelmK8sChangeAction = "change"
	HelmK8sActionChanged   HelmK8sChangeAction = "changed"
	HelmK8sActionDestroy   HelmK8sChangeAction = "destroy"
	HelmK8sActionDestroyed HelmK8sChangeAction = "destroyed"
)

// Summary holds counts for plan changes
type Summary struct {
	Create  int `json:"create"`
	Update  int `json:"update"`
	Delete  int `json:"delete"`
	Replace int `json:"replace"`
	Read    int `json:"read"`
	NoOp    int `json:"no-op"`
	Add     int `json:"add"`
	Change  int `json:"change"`
	Destroy int `json:"destroy"`
}

// TerraformPlan represents a Terraform plan structure
type TerraformPlan struct {
	ResourceDrift   []TerraformResourceDrift            `json:"resource_drift,omitempty"`
	ResourceChanges []TerraformResourceChange           `json:"resource_changes"`
	OutputChanges   map[string]TerraformOutputChangeRaw `json:"output_changes,omitempty"`
}

// TerraformResourceDrift represents a resource drift entry
type TerraformResourceDrift struct {
	Address       string                      `json:"address"`
	ModuleAddress *string                     `json:"module_address,omitempty"`
	Type          string                      `json:"type"`
	Name          string                      `json:"name"`
	Change        TerraformResourceChangeData `json:"change"`
}

// TerraformResourceChange represents a resource change entry
type TerraformResourceChange struct {
	Address       string                      `json:"address"`
	ModuleAddress *string                     `json:"module_address,omitempty"`
	Type          string                      `json:"type"`
	Name          string                      `json:"name"`
	Change        TerraformResourceChangeData `json:"change"`
}

// TerraformResourceChangeData holds the change details
type TerraformResourceChangeData struct {
	Actions      []TerraformChangeAction `json:"actions"`
	Before       any                     `json:"before,omitempty"`
	After        any                     `json:"after,omitempty"`
	AfterUnknown any                     `json:"after_unknown,omitempty"`
}

// TerraformOutputChangeRaw represents raw output change from JSON
type TerraformOutputChangeRaw struct {
	Actions         []TerraformChangeAction `json:"actions"`
	Before          any                     `json:"before,omitempty"`
	After           any                     `json:"after,omitempty"`
	AfterUnknown    any                     `json:"after_unknown,omitempty"`
	AfterSensitive  any                     `json:"after_sensitive,omitempty"`
	BeforeSensitive any                     `json:"before_sensitive,omitempty"`
}

// TerraformOutputChange represents a parsed output change
type TerraformOutputChange struct {
	Output          string                `json:"output"`
	Action          TerraformChangeAction `json:"action"`
	Before          any                   `json:"before,omitempty"`
	After           any                   `json:"after,omitempty"`
	AfterUnknown    any                   `json:"after_unknown,omitempty"`
	AfterSensitive  any                   `json:"after_sensitive,omitempty"`
	BeforeSensitive any                   `json:"before_sensitive,omitempty"`
}

// ParsedTerraformResourceChange represents a flattened resource change
type ParsedTerraformResourceChange struct {
	Address  string                `json:"address"`
	Module   *string               `json:"module,omitempty"`
	Resource string                `json:"resource"`
	Name     string                `json:"name"`
	Action   TerraformChangeAction `json:"action"`
	Before   any                   `json:"before,omitempty"`
	After    any                   `json:"after,omitempty"`
}

// HelmPlan represents a Helm plan structure
type HelmPlan struct {
	Plan            string         `json:"plan"`
	Op              string         `json:"op"`
	HelmContentDiff []HelmDiffItem `json:"helm_content_diff"`
}

// HelmDiffItem represents a single Helm diff item
type HelmDiffItem struct {
	API       string          `json:"api"`
	Kind      string          `json:"kind"`
	Name      string          `json:"name"`
	Namespace string          `json:"namespace"`
	Before    string          `json:"before,omitempty"`
	After     string          `json:"after,omitempty"`
	Entries   []HelmDiffEntry `json:"entries,omitempty"`
}

// HelmDiffEntry represents an entry in a Helm diff
type HelmDiffEntry struct {
	Path     string `json:"path"`
	Original string `json:"original"`
	Applied  string `json:"applied"`
	Type     int    `json:"type"`
	Payload  string `json:"payload"`
}

// ParsedHelmChange represents a parsed Helm change
type ParsedHelmChange struct {
	Workspace    string              `json:"workspace"`
	Release      string              `json:"release"`
	Resource     string              `json:"resource"`
	ResourceType string              `json:"resource_type"`
	Action       HelmK8sChangeAction `json:"action"`
	Before       *string             `json:"before,omitempty"`
	After        *string             `json:"after,omitempty"`
}

// KubernetesPlan represents a Kubernetes plan structure
type KubernetesPlan struct {
	Plan           string               `json:"plan"`
	Op             string               `json:"op"`
	K8sContentDiff []KubernetesDiffItem `json:"k8s_content_diff"`
}

// KubernetesDiffItem represents a single Kubernetes diff item
type KubernetesDiffItem struct {
	Version   string                `json:"_version"`
	Name      string                `json:"name"`
	Namespace string                `json:"namespace"`
	Kind      string                `json:"kind"`
	API       string                `json:"api"`
	Resource  string                `json:"resource"`
	Op        string                `json:"op"`
	Type      int                   `json:"type"` // 1=add, 2=delete, 3=change
	DryRun    bool                  `json:"dry_run"`
	Error     string                `json:"error,omitempty"`
	Entries   []KubernetesDiffEntry `json:"entries,omitempty"`
}

// KubernetesDiffEntry represents an entry in a Kubernetes diff
type KubernetesDiffEntry struct {
	Path     string `json:"path"`
	Original string `json:"original"`
	Applied  string `json:"applied"`
	Type     int    `json:"type"`
	Payload  string `json:"payload"`
}

// ParsedKubernetesChange represents a parsed Kubernetes change
type ParsedKubernetesChange struct {
	Namespace    string              `json:"namespace"`
	Name         string              `json:"name"`
	Resource     string              `json:"resource"`
	ResourceType string              `json:"resource_type"`
	Action       HelmK8sChangeAction `json:"action"`
	Before       *string             `json:"before,omitempty"`
	After        *string             `json:"after,omitempty"`
}

// ParsedKubernetesError represents a Kubernetes plan error
type ParsedKubernetesError struct {
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	Resource     string `json:"resource"`
	ResourceType string `json:"resource_type"`
	Error        string `json:"error"`
}

// ParsedTerraformPlan holds the parsed results of a Terraform plan
type ParsedTerraformPlan struct {
	Resources struct {
		Summary Summary                         `json:"summary"`
		Changes []ParsedTerraformResourceChange `json:"changes"`
	} `json:"resources"`
	Outputs struct {
		Summary Summary                 `json:"summary"`
		Changes []TerraformOutputChange `json:"changes"`
	} `json:"outputs"`
	Drift struct {
		Summary Summary                         `json:"summary"`
		Changes []ParsedTerraformResourceChange `json:"changes"`
	} `json:"drift"`
}

// ParsedHelmPlan holds the parsed results of a Helm plan
type ParsedHelmPlan struct {
	Summary Summary            `json:"summary"`
	Changes []ParsedHelmChange `json:"changes"`
}

// ParsedKubernetesPlan holds the parsed results of a Kubernetes plan
type ParsedKubernetesPlan struct {
	Summary Summary                  `json:"summary"`
	Changes []ParsedKubernetesChange `json:"changes"`
	Errors  []ParsedKubernetesError  `json:"errors,omitempty"`
}

// RunnerJobPlanWrapper represents the API response from GetRunnerJobPlan
type RunnerJobPlanWrapper struct {
	SandboxMode        *SandboxModePlan `json:"sandbox_mode,omitempty"`
	ApplyPlanContents  string           `json:"apply_plan_contents,omitempty"`
	ApplyPlanDisplay   string           `json:"apply_plan_display,omitempty"`
	Helm               *HelmModePlan    `json:"helm,omitempty"`
	Terraform          *TerraformMode   `json:"terraform,omitempty"`
	KubernetesManifest *K8sManifestMode `json:"kubernetes_manifest,omitempty"`
}

// SandboxModePlan holds nested plan data for sandbox mode
type SandboxModePlan struct {
	Helm               *HelmModePlan    `json:"helm,omitempty"`
	Terraform          *TerraformMode   `json:"terraform,omitempty"`
	KubernetesManifest *K8sManifestMode `json:"kubernetes_manifest,omitempty"`
}

// HelmModePlan holds Helm plan contents
type HelmModePlan struct {
	PlanContents string `json:"plan_contents,omitempty"`
}

// TerraformMode holds Terraform plan contents
type TerraformMode struct {
	PlanContents        string `json:"plan_contents,omitempty"`
	PlanDisplayContents string `json:"plan_display_contents,omitempty"`
}

// K8sManifestMode holds Kubernetes manifest plan contents
type K8sManifestMode struct {
	PlanContents string `json:"plan_contents,omitempty"`
}
