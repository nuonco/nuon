package credentials

type Config struct {
	ProjectID  string `json:"project_id" temporaljson:"project_id" hcl:"project_id" mapstructure:"project_id,omitempty" cty:"project_id"`
	Region     string `json:"region" temporaljson:"region" hcl:"region" mapstructure:"region,omitempty" cty:"region"`
	UseDefault bool   `json:"use_default" temporaljson:"use_default" hcl:"use_default,optional" mapstructure:"use_default,omitempty" cty:"use_default,optional"`
}

func (c Config) String() string {
	if c.UseDefault {
		return "default credentials"
	}

	return "service account"
}
