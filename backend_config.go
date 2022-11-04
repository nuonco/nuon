package terraform

import (
	"encoding/json"
)

const (
	backendConfigFilename string = "backend.json"
)

type backendConfigurer interface {
	createBackendConfig(BackendConfig, workspaceFileWriter) error
}

type BackendConfig struct {
	BucketName   string `json:"bucket"`
	BucketKey    string `json:"key"`
	BucketRegion string `json:"region"`
}

type workspaceFileWriter interface {
	writeFile(string, []byte) error
}

var _ backendConfigurer = (*s3BackendConfigurer)(nil)

type s3BackendConfigurer struct{}

func (t *s3BackendConfigurer) createBackendConfig(cfg BackendConfig, wkspace workspaceFileWriter) error {
	byts, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return wkspace.writeFile(backendConfigFilename, byts)
}
