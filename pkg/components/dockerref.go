package components

import "encoding/json"

type DockerRef struct {
	Name  string `json:"-"`
	Image string
	Tag   string
}

func (dr *DockerRef) GetName() string {
	return dr.Name
}

func (dr DockerRef) ToJSON() ([]byte, error) {
	contentJSON, err := json.Marshal(dr)
	if err != nil {
		return nil, err
	}

	content := map[string][]byte{
		dr.GetName(): contentJSON,
	}

	return json.Marshal(content)
}
