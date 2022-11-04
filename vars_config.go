package terraform

import (
	"encoding/json"
)

const (
	varsConfigFilename string = "nuon.tfvars.json"
)

type varsConfigurer interface {
	createVarsConfigFile(map[string]interface{}, workspaceFileWriter) error
}

var _ varsConfigurer = (*tfVarsConfigurer)(nil)

type tfVarsConfigurer struct{}

func (t *tfVarsConfigurer) createVarsConfigFile(cfg map[string]interface{}, wkspace workspaceFileWriter) error {
	byts, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return wkspace.writeFile(varsConfigFilename, byts)
}
