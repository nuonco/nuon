package components

import "encoding/json"

type Kubernetes struct {
	Name string `json:"-"`
}

func (kb *Kubernetes) GetName() string {
	return kb.Name
}

func (kb *Kubernetes) ToJSON() ([]byte, error) {
	contentJSON, err := json.Marshal(kb)
	if err != nil {
		return nil, err
	}

	content := map[string][]byte{
		kb.GetName(): contentJSON,
	}

	return json.Marshal(content)
}
