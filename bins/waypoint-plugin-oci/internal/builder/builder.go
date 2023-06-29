package builder

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"oras.land/oras-go/v2/content/file"
)

var _ component.Builder = (*Builder)(nil)

type Builder struct {
	v   *validator.Validate
	cfg configs.OCIArchiveBuild

	// fields set by the plugin execution
	tmpDir   string
	chartDir string
	store    *file.Store
}

func New(v *validator.Validate) (*Builder, error) {
	tmpDir, err := ioutil.TempDir("", "helm-package-push")
	if err != nil {
		return nil, fmt.Errorf("unable to load temp dir: %s", err)
	}

	storeDir := filepath.Join(tmpDir, "store")
	store, err := file.New(storeDir)
	if err != nil {
		return nil, fmt.Errorf("unable to get file store: %w", err)
	}

	return &Builder{
		v:      v,
		tmpDir: tmpDir,
		store:  store,
	}, nil
}
