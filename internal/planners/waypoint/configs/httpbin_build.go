package configs

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/go-playground/validator/v10"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

const (
	httpbinBuildTmplName string = "httpbin-build"
)

// NewHttpbinBuildBuilder returns a builder that renders our hardcoded sample application
func NewHttpbinBuildBuilder(v *validator.Validate, opts ...baseBuilderOption) (*httpbinBuildBuilder, error) {
	baseBuilder, err := newBaseBuilder(v, opts...)
	if err != nil {
		return nil, err
	}
	return &httpbinBuildBuilder{baseBuilder}, nil
}

type httpbinBuildBuilder struct {
	*baseBuilder
}

var _ Builder = (*httpbinBuildBuilder)(nil)

var httpbinBuildTmpl string = `
project = "{{.WaypointRef.Project}}"

app "{{.WaypointRef.App}}" {
  build {
    use "docker-pull" {
      image = "kennethreitz/httpbin"
      tag   = "latest"
    }

    registry {
      use "aws-ecr" {
	repository = "{{.EcrRef.RepositoryName}}"
	tag	 = "{{.EcrRef.Tag}}"
	region = "{{.EcrRef.Region}}"
      }
    }
  }

  deploy {
    use "kubernetes" {}
  }
}
`

func (s *httpbinBuildBuilder) Render() ([]byte, waypointv1.Hcl_Format, error) {
	tmpl, err := template.New(httpbinBuildTmplName).Parse(httpbinBuildTmpl)
	if err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to parse static template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), waypointv1.Hcl_HCL, nil
}
