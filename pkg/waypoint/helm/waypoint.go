package helm

const (
	waypointImageRepository string = "public.ecr.aws/p7e3r5y0/waypoint"
	waypointVersion         string = "v0.1.0"
)

// Values represent all of the possible values for a helm installation
type Values struct {
	Server struct {
		Enabled bool `mapstructure:"enabled"`

		Image struct {
			Repository string `mapstructure:"repository,omitempty"`
			Tag        string `mapstructure:"tag,omitempty"`
		} `mapstructure:"image,omitempty"`

		RunArgs []string `mapstructure:"runArgs,omitempty"`
		Domain  string   `mapstructure:"domain,omitempty"`
		Certs   struct {
			SecretName        interface{} `mapstructure:"secretName,omitempty"`
			CertName          string      `mapstructure:"certName,omitempty"`
			KeyName           string      `mapstructure:"keyName,omitempty"`
			ClusterIssuerName string      `mapstructure:"clusterIssuerName,omitempty"`
		} `mapstructure:"certs,omitempty"`

		Storage struct {
			Size         string            `mapstructure:"size,omitempty"`
			StorageClass interface{}       `mapstructure:"storageClass,omitempty"`
			Annotations  map[string]string `mapstructure:"annotations,omitempty"`
		} `mapstructure:"storage,omitempty"`

		Resources struct {
			Requests struct {
				Memory string `mapstructure:"memory,omitempty"`
				CPU    string `mapstructure:"cpu,omitempty"`
			} `mapstructure:"requests,omitempty"`
		} `mapstructure:"resources,omitempty"`

		PriorityClassName string            `mapstructure:"priorityClassName,omitempty"`
		ExtraLabels       map[string]string `mapstructure:"extraLabels,omitempty"`
		Annotations       map[string]string `mapstructure:"annotations,omitempty"`
		NodeSelector      interface{}       `mapstructure:"nodeSelector,omitempty"`
		ServiceAccount    struct {
			Create      bool              `mapstructure:"create"`
			Name        string            `mapstructure:"name,omitempty"`
			Annotations map[string]string `mapstructure:"annotations,omitempty"`
		} `mapstructure:"serviceAccount,omitempty"`
		StatefulSet struct {
			Annotations map[string]string `mapstructure:"annotations,omitempty"`
		} `mapstructure:"statefulSet,omitempty"`
		Service struct {
			Annotations map[string]string `mapstructure:"annotations,omitempty"`
		} `mapstructure:"service,omitempty"`
	} `mapstructure:"server,omitempty"`

	Runner struct {
		Enabled  bool   `mapstructure:"enabled"`
		ID       string `mapstructure:"id,omitempty"`
		Replicas int    `mapstructure:"replicas,omitempty"`
		Image    struct {
			Repository string `mapstructure:"repository,omitempty"`
			Tag        string `mapstructure:"tag,omitempty"`
		} `mapstructure:"image,omitempty"`
		AgentArgs []string `mapstructure:"agentArgs,omitempty"`
		Server    struct {
			Addr          string `mapstructure:"addr,omitempty"`
			TLS           bool   `mapstructure:"tls,omitempty"`
			TLSSkipVerify bool   `mapstructure:"tlsSkipVerify,omitempty"`
			Cookie        string `mapstructure:"cookie,omitempty"`
			TokenSecret   string `mapstructure:"tokenSecret,omitempty"`
		} `mapstructure:"server,omitempty"`
		Storage struct {
			Size         string      `mapstructure:"size,omitempty"`
			StorageClass interface{} `mapstructure:"storageClass,omitempty"`
		} `mapstructure:"storage,omitempty"`
		Odr struct {
			Image struct {
				Repository string `mapstructure:"repository,omitempty"`
				Tag        string `mapstructure:"tag,omitempty"`
			} `mapstructure:"image,omitempty"`
			ManagedNamespaces []interface{} `mapstructure:"managedNamespaces,omitempty"`
			ServiceAccount    struct {
				Create      bool              `mapstructure:"create,omitempty"`
				Name        string            `mapstructure:"name,omitempty"`
				Annotations map[string]string `mapstructure:"annotations,omitempty"`
			} `mapstructure:"serviceAccount,omitempty"`
		} `mapstructure:"odr,omitempty"`
		Resources struct {
			Requests struct {
				Memory string `mapstructure:"memory,omitempty"`
				CPU    string `mapstructure:"cpu,omitempty"`
			} `mapstructure:"requests,omitempty"`
		} `mapstructure:"resources,omitempty"`
		PriorityClassName string            `mapstructure:"priorityClassName,omitempty"`
		ExtraLabels       map[string]string `mapstructure:"extraLabels,omitempty"`
		Annotations       map[string]string `mapstructure:"annotations,omitempty"`
		NodeSelector      interface{}       `mapstructure:"nodeSelector,omitempty"`
		ServiceAccount    struct {
			Create      bool              `mapstructure:"create"`
			Name        string            `mapstructure:"name,omitempty"`
			Annotations map[string]string `mapstructure:"annotations,omitempty"`
		} `mapstructure:"serviceAccount,omitempty"`
		Deployment struct {
			Annotations map[string]string `mapstructure:"annotations,omitempty"`
		} `mapstructure:"deployment,omitempty"`
	} `mapstructure:"runner,omitempty"`
	UI struct {
		Service struct {
			Enabled        bool              `mapstructure:"enabled"`
			Type           string            `mapstructure:"type,omitempty"`
			Annotations    map[string]string `mapstructure:"annotations,omitempty"`
			AdditionalSpec interface{}       `mapstructure:"additionalSpec,omitempty"`
		} `mapstructure:"service,omitempty"`
		Ingress struct {
			Enabled bool `mapstructure:"enabled,omitempty"`
			Hosts   []struct {
				Host  string        `mapstructure:"host,omitempty"`
				Paths []interface{} `mapstructure:"paths,omitempty"`
			} `mapstructure:"hosts,omitempty"`
			Labels      map[string]string `mapstructure:"labels,omitempty"`
			Annotations map[string]string `mapstructure:"annotations,omitempty"`
			ExtraPaths  []interface{}     `mapstructure:"extraPaths,omitempty"`
			TLS         []interface{}     `mapstructure:"tls,omitempty"`
		} `mapstructure:"ingress,omitempty"`
	} `mapstructure:"ui,omitempty"`
	Bootstrap struct {
		ServiceAccount struct {
			Create      bool              `mapstructure:"create"`
			Name        string            `mapstructure:"name,omitempty"`
			Annotations map[string]string `mapstructure:"annotations,omitempty"`
		} `mapstructure:"serviceAccount,omitempty"`
	} `mapstructure:"bootstrap,omitempty"`
}

// NewDefaultInstallValues returns a values set with defaults for installing a runner in a sandbox
func NewDefaultInstallValues() *Values {
	var vals Values
	vals.Runner.Enabled = true
	vals.Runner.Replicas = 1
	vals.Runner.Server.TLS = true
	vals.Runner.Server.TLSSkipVerify = true
	vals.Runner.Storage.Size = "1Gi"
	vals.Runner.Odr.ServiceAccount.Create = true
	vals.Runner.Resources.Requests.Memory = "256Mi"
	vals.Runner.Resources.Requests.CPU = "250m"
	vals.Runner.Image.Repository = waypointImageRepository
	vals.Runner.Image.Tag = waypointVersion

	return &vals
}

// NewDefaultInstallValues returns a values set with defaults for installing a runner in a sandbox
func NewDefaultOrgRunnerValues() *Values {
	var vals Values
	vals.Server.Enabled = false
	vals.Server.ServiceAccount.Create = false

	vals.Runner.Enabled = true
	vals.Runner.Replicas = 1
	vals.Runner.Server.TLS = true
	vals.Runner.Server.TLSSkipVerify = true
	vals.Runner.Storage.Size = "1Gi"
	vals.Runner.Odr.ServiceAccount.Create = true
	vals.Runner.Resources.Requests.Memory = "256Mi"
	vals.Runner.Resources.Requests.CPU = "250m"
	vals.Runner.Image.Repository = waypointImageRepository
	vals.Runner.Image.Tag = waypointVersion

	vals.Bootstrap.ServiceAccount.Create = false

	return &vals
}

// NewDefaultOrgServerValues returns a values set with defaults for installing a runner, server and UI
func NewDefaultOrgServerValues() *Values {
	var vals Values

	vals.Server.Enabled = true
	vals.Server.Image.Repository = waypointImageRepository
	vals.Server.Image.Tag = waypointVersion

	vals.Runner.Enabled = false
	vals.Runner.ServiceAccount.Create = false

	vals.UI.Service.Enabled = true
	vals.UI.Service.Type = "ClusterIP"
	vals.Bootstrap.ServiceAccount.Create = false

	return &vals
}
