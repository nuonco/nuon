package builder

import (
	"fmt"
	"io/ioutil"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

var _ component.Builder = (*Builder)(nil)

type Builder struct {
	v   *validator.Validate
	cfg configs.OCIArchiveBuild

	// fields set by the plugin execution
	tmpDir   string
	chartDir string
}

func New(v *validator.Validate) (*Builder, error) {
	tmpDir, err := ioutil.TempDir("", "helm-package-push")
	if err != nil {
		return nil, fmt.Errorf("unable to load temp dir: %s", err)
	}

	return &Builder{
		v:      v,
		tmpDir: tmpDir,
	}, nil
}
