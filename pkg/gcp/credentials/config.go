package credentials

import "encoding/json"

type Config struct {
	ProjectID                 string `json:"project_id" temporaljson:"project_id"`
	Region                    string `json:"region" temporaljson:"region"`
	ImpersonateServiceAccount string `json:"impersonate_service_account,omitempty" temporaljson:"impersonate_service_account,omitempty"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	var b bool
	if json.Unmarshal(data, &b) == nil {
		return nil
	}

	type config Config
	return json.Unmarshal(data, (*config)(c))
}
