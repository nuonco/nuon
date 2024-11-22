package vars

type instanceImageRepository struct {
	ID          string `json:"id,omitempty"`
	ARN         string `json:"arn,omitempty"`
	Name        string `json:"name,omitempty"`
	URI         string `json:"uri,omitempty"`
	Image       string `json:"image,omitempty"`
	LoginServer string `json:"login_server,omitempty"`
}

type instanceImageRegistry struct {
	ID string `json:"id"`
}

type instanceImage struct {
	Tag        string                  `json:"tag"`
	Repository instanceImageRepository `json:"repository"`
	Registry   instanceImageRegistry   `json:"registry"`
}

type instanceIntermediate struct {
	Image   instanceImage          `json:"image"`
	Outputs map[string]interface{} `json:"outputs"`
}

type appIntermediate struct {
	ID      string            `json:"id"`
	Secrets map[string]string `json:"secrets" faker:"-"`
}

type orgIntermediate struct {
	ID string `json:"id"`
}

type installSandboxIntermediate struct {
	Type    string                 `json:"type"`
	Version string                 `json:"version"`
	Outputs map[string]interface{} `json:"outputs" faker:"-"`
}

type installIntermediate struct {
	ID string `json:"id"`

	PublicDomain   string                     `json:"public_domain"`
	InternalDomain string                     `json:"internal_domain"`
	Sandbox        installSandboxIntermediate `json:"sandbox"`
	Inputs         map[string]string          `json:"inputs"`
}

type imageIntermediate struct {
	Tag        string `json:"tag"`
	Repository string `json:"repository"`
}

type componentIntermediate struct {
	Outputs map[string]string `json:"outputs"`
	Image   imageIntermediate `json:"image"`
}

// intermediate represents the intermediate data available to users to interpolate
type intermediate struct {
	Org        orgIntermediate                  `json:"org"`
	App        appIntermediate                  `json:"app"`
	Install    installIntermediate              `json:"install"`
	Components map[string]*instanceIntermediate `json:"components" faker:"-"`
}
