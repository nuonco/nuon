package configs

type BasicDeploy struct {
	Plugin string `hcl:"plugin,label"`

	Annotations map[string]string `hcl:"annotations,optional"`

	AutoscaleConfig *AutoscaleConfig `hcl:"autoscale,block"`

	Context string `hcl:"context,optional"`

	Count int32 `hcl:"replicas,optional"`

	ImageSecret string `hcl:"image_secret,optional"`

	KubeconfigPath string `hcl:"kubeconfig,optional"`

	Labels map[string]string `hcl:"labels,optional"`

	Namespace string `hcl:"namespace,optional"`

	ProbePath string `hcl:"probe_path,optional"`

	Probe *Probe `hcl:"probe,block"`

	Resources map[string]string `hcl:"resources,optional"`

	CPU *ResourceConfig `hcl:"cpu,block"`

	Memory *ResourceConfig `hcl:"memory,block"`

	ScratchSpace []string `hcl:"scratch_path,optional"`

	ServiceAccount string `hcl:"service_account,optional"`

	ServicePort *uint `hcl:"service_port,optional"`

	StaticEnvVars map[string]string `hcl:"static_environment,optional"`

	Pod *Pod `hcl:"pod,block"`

	DeprecatedPorts []map[string]string `hcl:"ports,optional" docs:"hidden"`
}

type ResourceConfig struct {
	Request string `hcl:"request,optional" json:"request"`
	Limit   string `hcl:"limit,optional" json:"limit"`
}

type AutoscaleConfig struct {
	MinReplicas int32 `hcl:"min_replicas,optional"`
	MaxReplicas int32 `hcl:"max_replicas,optional"`
	TargetCPU   int32 `hcl:"cpu_percent,optional"`
}

type Pod struct {
	SecurityContext *PodSecurityContext `hcl:"security_context,block"`
	Container       *Container          `hcl:"container,block"`
	Sidecars        []*Sidecar          `hcl:"sidecar,block"`
}

type Sidecar struct {
	Image string `hcl:"image"`

	Container *Container `hcl:"container,block"`
}

type Port struct {
	Name     string `hcl:"name"`
	Port     uint   `hcl:"port"`
	HostPort uint   `hcl:"host_port,optional"`
	HostIP   string `hcl:"host_ip,optional"`
	Protocol string `hcl:"protocol,optional"`
}

type Container struct {
	Name          string            `hcl:"name,optional"`
	Ports         []*Port           `hcl:"port,block"`
	ProbePath     string            `hcl:"probe_path,optional"`
	Probe         *Probe            `hcl:"probe,block"`
	CPU           *ResourceConfig   `hcl:"cpu,block"`
	Memory        *ResourceConfig   `hcl:"memory,block"`
	Resources     map[string]string `hcl:"resources,optional"`
	Command       *[]string         `hcl:"command,optional"`
	Args          *[]string         `hcl:"args,optional"`
	StaticEnvVars map[string]string `hcl:"static_environment,optional"`
}

type PodSecurityContext struct {
	RunAsUser    *int64 `hcl:"run_as_user"`
	RunAsGroup   *int64 `hcl:"run_as_group"`
	RunAsNonRoot *bool  `hcl:"run_as_non_root"`
	FsGroup      *int64 `hcl:"fs_group"`
}

type Probe struct {
	InitialDelaySeconds uint `hcl:"initial_delay,optional"`

	TimeoutSeconds uint `hcl:"timeout,optional"`

	FailureThreshold uint `hcl:"failure_threshold,optional"`
}
