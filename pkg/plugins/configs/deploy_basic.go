package configs

type BasicDeploy struct {
	Plugin string `hcl:"plugin,label"`

	ServicePort int `hcl:"service_port"`
}
