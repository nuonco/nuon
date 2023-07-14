package configs

type BasicDeploy struct {
	Plugin string `hcl:"plugin,label"`

	// Annotations are added to the pod spec of the deployed application.  This is
	// useful when using mutating webhook admission controllers to further process
	// pod events.
	Annotations map[string]string `hcl:"annotations,optional"`

	// AutoscaleConfig will create a horizontal pod autoscaler for a given
	// deployment and scale the replica pods up or down based on a given
	// load metric, such as CPU utilization
	AutoscaleConfig *AutoscaleConfig `hcl:"autoscale,block"`

	// Context specifies the kube context to use.
	Context string `hcl:"context,optional"`

	// The number of replicas of the service to maintain. If this number is maintained
	// outside waypoint, for instance by a pod autoscaler, do not set this variable.
	Count int32 `hcl:"replicas,optional"`

	// The name of the Kubernetes secret to use to pull the image stored
	// in the registry.
	// TODO This maybe should be required because the vast majority of deployments
	// will be against private images.
	ImageSecret string `hcl:"image_secret,optional"`

	// KubeconfigPath is the path to the kubeconfig file. If this is
	// blank then we default to the home directory.
	KubeconfigPath string `hcl:"kubeconfig,optional"`

	// A map of key vals to label the deployed Pod and Deployment with.
	Labels map[string]string `hcl:"labels,optional"`

	// Namespace is the Kubernetes namespace to target the deployment to.
	Namespace string `hcl:"namespace,optional"`

	// If set, this is the HTTP path to request to test that the application
	// is up and running. Without this, we only test that a connection can be
	// made to the port.
	ProbePath string `hcl:"probe_path,optional"`

	// Probe details for describing a health check to be performed against a container.
	Probe *Probe `hcl:"probe,block"`

	// Optionally define various resources limits for kubernetes pod containers
	// such as memory and cpu.
	Resources map[string]string `hcl:"resources,optional"`

	// Optionally define various cpu resource limits and requests for kubernetes pod containers
	CPU *ResourceConfig `hcl:"cpu,block"`

	// Optionally define various memory resource limits and requests for kubernetes pod containers
	Memory *ResourceConfig `hcl:"memory,block"`

	// An array of paths to directories that will be mounted as EmptyDirVolumes in the pod
	// to store temporary data.
	ScratchSpace []string `hcl:"scratch_path,optional"`

	// ServiceAccount is the name of the Kubernetes service account to apply to the
	// application deployment. This is useful to apply Kubernetes RBAC to the pod.
	ServiceAccount string `hcl:"service_account,optional"`

	// Port that your service is running on within the actual container.
	// Defaults to DefaultServicePort const.
	// NOTE: Ports and ServicePort cannot both be defined
	ServicePort *uint `hcl:"service_port,optional"`

	// Environment variables that are meant to configure the application in a static
	// way. This might be control an image that has multiple modes of operation,
	// selected via environment variable. Most configuration should use the waypoint
	// config commands.
	StaticEnvVars map[string]string `hcl:"static_environment,optional"`

	// Pod describes the configuration for the pod
	Pod *Pod `hcl:"pod,block"`

	// Deprecated field, previous definition of ports
	DeprecatedPorts []map[string]string `hcl:"ports,optional" docs:"hidden"`
}

// ResourceConfig describes the request and limit of a resource. Used for
// cpu and memory resource configuration.
type ResourceConfig struct {
	Request string `hcl:"request,optional" json:"request"`
	Limit   string `hcl:"limit,optional" json:"limit"`
}

// AutoscaleConfig describes the possible configuration for creating a
// horizontal pod autoscaler
type AutoscaleConfig struct {
	MinReplicas int32 `hcl:"min_replicas,optional"`
	MaxReplicas int32 `hcl:"max_replicas,optional"`
	// TargetCPU will determine the max load before the autoscaler will increase
	// a replica
	TargetCPU int32 `hcl:"cpu_percent,optional"`
}

// Pod describes the configuration for the pod
type Pod struct {
	SecurityContext *PodSecurityContext `hcl:"security_context,block"`
	Container       *Container          `hcl:"container,block"`
	Sidecars        []*Sidecar          `hcl:"sidecar,block"`
}

type Sidecar struct {
	// Specifying Image in Container would make it visible on the main Pod config,
	// which isn't the right way to specify the app image.
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

// Container describes the detailed parameters to declare a kubernetes container
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

// PodSecurityContext describes the security config for the Pod
type PodSecurityContext struct {
	RunAsUser    *int64 `hcl:"run_as_user"`
	RunAsGroup   *int64 `hcl:"run_as_group"`
	RunAsNonRoot *bool  `hcl:"run_as_non_root"`
	FsGroup      *int64 `hcl:"fs_group"`
}

// Probe describes a health check to be performed against a container to determine whether it is
// alive or ready to receive traffic.
type Probe struct {
	// Time in seconds to wait before performing the initial liveness and readiness probes.
	// Defaults to 5 seconds.
	InitialDelaySeconds uint `hcl:"initial_delay,optional"`

	// Time in seconds before the probe fails.
	// Defaults to 5 seconds.
	TimeoutSeconds uint `hcl:"timeout,optional"`

	// Number of times a liveness probe can fail before the container is killed.
	// FailureThreshold * TimeoutSeconds should be long enough to cover your worst
	// case startup times. Defaults to 30 failures.
	FailureThreshold uint `hcl:"failure_threshold,optional"`
}
