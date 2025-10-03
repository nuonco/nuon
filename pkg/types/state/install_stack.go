package state

func NewInstallStackState() *InstallStackState {
	return &InstallStackState{
		Outputs: make(map[string]interface{}, 0),
	}
}

type InstallStackState struct {
	Populated bool `json:"populated"`

	QuickLinkURL string `json:"quick_link_url"`
	TemplateURL  string `json:"template_url"`
	TemplateJSON string `json:"template_json"`
	Checksum     string `json:"checksum"`
	Status       string `json:"status"`

	Outputs map[string]interface{} `json:"outputs"`
}
