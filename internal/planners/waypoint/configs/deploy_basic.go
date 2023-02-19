package configs

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/go-playground/validator/v10"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

const (
	basicDeployTmplName string = "basic-deploy"
)

// NewBasicDeploy returns a builder that creates a configuration for a basic deployment.
func NewBasicDeploy(v *validator.Validate, opts ...Option) (*basicDeploy, error) {
	baseBuilder, err := newBaseBuilder(v, opts...)
	if err != nil {
		return nil, err
	}
	return &basicDeploy{baseBuilder}, nil
}

type basicDeploy struct {
	*baseBuilder
}

var _ Builder = (*basicDeploy)(nil)

var basicDeployTmpl string = `
project = "{{.WaypointRef.Project}}"

app "{{.WaypointRef.App}}" {
  build {
    registry {
      use "aws-ecr" {
	repository = "{{.EcrRef.RepositoryName}}"
	tag	 = "{{.EcrRef.Tag}}"
      }
    }
  }

  deploy {
    use "kubernetes" {
      service_port = 80
    }
  }
}
`

func (s *basicDeploy) Render() ([]byte, waypointv1.Hcl_Format, error) {
	tmpl, err := template.New(basicDeployTmplName).Parse(basicDeployTmpl)
	if err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to parse static template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), waypointv1.Hcl_HCL, nil
}
