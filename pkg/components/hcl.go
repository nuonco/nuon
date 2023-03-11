package components

import "encoding/json"

type HCL struct {
	Project string
	App     *App
}

func (hcl *HCL) ToJSON() ([]byte, error) {
	appContent, err := hcl.App.ToJSON()
	if err != nil {
		return nil, err
	}

	content := map[string][]byte{
		"project": []byte(hcl.Project),
		"app":     appContent,
	}

	return json.Marshal(content)
}
