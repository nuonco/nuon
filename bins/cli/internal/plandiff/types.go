package plandiff

type PlanType string

const (
	PlanTypeTerraform  PlanType = "terraform_plan"
	PlanTypeHelm       PlanType = "helm_approval"
	PlanTypeKubernetes PlanType = "kubernetes_manifest_approval"
	PlanTypeUnknown    PlanType = "unknown"
)

type TerraformChangeAction string

const (
	TerraformActionCreate  TerraformChangeAction = "create"
	TerraformActionUpdate  TerraformChangeAction = "update"
	TerraformActionDelete  TerraformChangeAction = "delete"
	TerraformActionNoOp    TerraformChangeAction = "no-op"
	TerraformActionReplace TerraformChangeAction = "replace"
	TerraformActionRead    TerraformChangeAction = "read"
)

type HelmK8sChangeAction string

const (
	HelmK8sActionAdd       HelmK8sChangeAction = "add"
	HelmK8sActionAdded     HelmK8sChangeAction = "added"
	HelmK8sActionChange    HelmK8sChangeAction = "change"
	HelmK8sActionChanged   HelmK8sChangeAction = "changed"
	HelmK8sActionDestroy   HelmK8sChangeAction = "destroy"
	HelmK8sActionDestroyed HelmK8sChangeAction = "destroyed"
)

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

type TerraformPlan struct {
	ResourceDrift   []TerraformResourceDrift            `json:"resource_drift,omitempty"`
	ResourceChanges []TerraformResourceChange           `json:"resource_changes"`
	OutputChanges   map[string]TerraformOutputChangeRaw `json:"output_changes,omitempty"`
}

type TerraformResourceDrift struct {
	Address       string                      `json:"address"`
	ModuleAddress *string                     `json:"module_address,omitempty"`
	Type          string                      `json:"type"`
	Name          string                      `json:"name"`
	Change        TerraformResourceChangeData `json:"change"`
}

type TerraformResourceChange struct {
	Address       string                      `json:"address"`
	ModuleAddress *string                     `json:"module_address,omitempty"`
	Type          string                      `json:"type"`
	Name          string                      `json:"name"`
	Change        TerraformResourceChangeData `json:"change"`
}

type TerraformResourceChangeData struct {
	Actions      []TerraformChangeAction `json:"actions"`
	Before       any                     `json:"before,omitempty"`
	After        any                     `json:"after,omitempty"`
	AfterUnknown any                     `json:"after_unknown,omitempty"`
}

type TerraformOutputChangeRaw struct {
	Actions         []TerraformChangeAction `json:"actions"`
	Before          any                     `json:"before,omitempty"`
	After           any                     `json:"after,omitempty"`
	AfterUnknown    any                     `json:"after_unknown,omitempty"`
	AfterSensitive  any                     `json:"after_sensitive,omitempty"`
	BeforeSensitive any                     `json:"before_sensitive,omitempty"`
}

type TerraformOutputChange struct {
	Output          string                `json:"output"`
	Action          TerraformChangeAction `json:"action"`
	Before          any                   `json:"before,omitempty"`
	After           any                   `json:"after,omitempty"`
	AfterUnknown    any                   `json:"after_unknown,omitempty"`
	AfterSensitive  any                   `json:"after_sensitive,omitempty"`
	BeforeSensitive any                   `json:"before_sensitive,omitempty"`
}

type ParsedTerraformResourceChange struct {
	Address  string                `json:"address"`
	Module   *string               `json:"module,omitempty"`
	Resource string                `json:"resource"`
	Name     string                `json:"name"`
	Action   TerraformChangeAction `json:"action"`
	Before   any                   `json:"before,omitempty"`
	After    any                   `json:"after,omitempty"`
}

type HelmPlan struct {
	Plan            string         `json:"plan"`
	Op              string         `json:"op"`
	HelmContentDiff []HelmDiffItem `json:"helm_content_diff"`
}

type HelmDiffItem struct {
	API       string          `json:"api"`
	Kind      string          `json:"kind"`
	Name      string          `json:"name"`
	Namespace string          `json:"namespace"`
	Before    string          `json:"before,omitempty"`
	After     string          `json:"after,omitempty"`
	Entries   []HelmDiffEntry `json:"entries,omitempty"`
}

type HelmDiffEntry struct {
	Path     string `json:"path"`
	Original string `json:"original"`
	Applied  string `json:"applied"`
	Type     int    `json:"type"`
	Payload  string `json:"payload"`
}

type ParsedHelmChange struct {
	Workspace    string              `json:"workspace"`
	Release      string              `json:"release"`
	Resource     string              `json:"resource"`
	ResourceType string              `json:"resource_type"`
	Action       HelmK8sChangeAction `json:"action"`
	Before       *string             `json:"before,omitempty"`
	After        *string             `json:"after,omitempty"`
}

type KubernetesPlan struct {
	Plan           string               `json:"plan"`
	Op             string               `json:"op"`
	K8sContentDiff []KubernetesDiffItem `json:"k8s_content_diff"`
}

type KubernetesDiffItem struct {
	Version   string                `json:"_version"`
	Name      string                `json:"name"`
	Namespace string                `json:"namespace"`
	Kind      string                `json:"kind"`
	API       string                `json:"api"`
	Resource  string                `json:"resource"`
	Op        string                `json:"op"`
	Type      int                   `json:"type"`
	DryRun    bool                  `json:"dry_run"`
	Error     string                `json:"error,omitempty"`
	Entries   []KubernetesDiffEntry `json:"entries,omitempty"`
}

type KubernetesDiffEntry struct {
	Path     string `json:"path"`
	Original string `json:"original"`
	Applied  string `json:"applied"`
	Type     int    `json:"type"`
	Payload  string `json:"payload"`
}

type ParsedKubernetesChange struct {
	Namespace    string              `json:"namespace"`
	Name         string              `json:"name"`
	Resource     string              `json:"resource"`
	ResourceType string              `json:"resource_type"`
	Action       HelmK8sChangeAction `json:"action"`
	Before       *string             `json:"before,omitempty"`
	After        *string             `json:"after,omitempty"`
}

type ParsedKubernetesError struct {
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	Resource     string `json:"resource"`
	ResourceType string `json:"resource_type"`
	Error        string `json:"error"`
}

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

type ParsedHelmPlan struct {
	Summary Summary            `json:"summary"`
	Changes []ParsedHelmChange `json:"changes"`
}

type ParsedKubernetesPlan struct {
	Summary Summary                  `json:"summary"`
	Changes []ParsedKubernetesChange `json:"changes"`
	Errors  []ParsedKubernetesError  `json:"errors,omitempty"`
}

type RunnerJobPlanWrapper struct {
	SandboxMode        *SandboxModePlan `json:"sandbox_mode,omitempty"`
	ApplyPlanContents  string           `json:"apply_plan_contents,omitempty"`
	ApplyPlanDisplay   string           `json:"apply_plan_display,omitempty"`
	Helm               *HelmModePlan    `json:"helm,omitempty"`
	Terraform          *TerraformMode   `json:"terraform,omitempty"`
	KubernetesManifest *K8sManifestMode `json:"kubernetes_manifest,omitempty"`
}

type SandboxModePlan struct {
	Helm               *HelmModePlan    `json:"helm,omitempty"`
	Terraform          *TerraformMode   `json:"terraform,omitempty"`
	KubernetesManifest *K8sManifestMode `json:"kubernetes_manifest,omitempty"`
}

type HelmModePlan struct {
	PlanContents string `json:"plan_contents,omitempty"`
}

type TerraformMode struct {
	PlanContents        string `json:"plan_contents,omitempty"`
	PlanDisplayContents string `json:"plan_display_contents,omitempty"`
}

type K8sManifestMode struct {
	PlanContents string `json:"plan_contents,omitempty"`
}
