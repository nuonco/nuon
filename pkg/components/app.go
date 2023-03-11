package components

import "encoding/json"

type App struct {
	Name   string `json:"-"`
	Build  *UseBlock
	Deploy *UseBlock
}

func (a *App) ToJSON() ([]byte, error) {
	buildJSON, err := a.Build.ToJSON()
	if err != nil {
		return nil, err
	}

	deployJSON, err := a.Deploy.ToJSON()
	if err != nil {
		return nil, err
	}

	content := map[string][]byte{
		"build":  buildJSON,
		"deploy": deployJSON,
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	appContent := map[string][]byte{
		a.Name: contentJSON,
	}

	return json.Marshal(appContent)
}

type Operation interface {
	ToJSON() ([]byte, error)
}

type UseBlock struct {
	Use Operation
}

func (ub *UseBlock) ToJSON() ([]byte, error) {
	return ub.Use.ToJSON()
}
