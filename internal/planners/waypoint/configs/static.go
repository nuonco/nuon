package configs

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/go-playground/validator/v10"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

// NewStaticBuildConfig returns a builder that renders our hardcoded sample application
func NewStaticBuilder(v *validator.Validate, opts ...baseBuilderOption) (*staticBuilder, error) {
	baseBuilder, err := newBaseBuilder(v, opts...)
	if err != nil {
		return nil, err
	}
	return &staticBuilder{baseBuilder}, nil
}

type staticBuilder struct {
	*baseBuilder
}

var _ Builder = (*staticBuilder)(nil)

var staticBuildTmpl string = `
project = "{{.Project}}"

app "{{.AppName}}" {
  build {
    use "docker-pull" {
      image = "{{.InputImage}}"
      tag   = "{{.InputVersion}}"
    }

    registry {
      use "aws-ecr" {
	repository = "{{.OutputRepository}}"
	tag	 = "{{.OutputVersion}}"
	region = "us-west-2"
      }
    }
  }

  deploy {
    use "kubernetes" {}
  }
}
`

type staticTmplArgs struct {
	Project          string
	AppName          string
	InputImage       string
	InputVersion     string
	OutputRepository string
	OutputVersion    string
}

func (s *staticBuilder) Render() ([]byte, waypointv1.Hcl_Format, error) {
	tmpl, err := template.New("build-config").Parse(staticBuildTmpl)
	if err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to parse static template: %w", err)
	}

	buf := new(bytes.Buffer)
	args := staticTmplArgs{
		Project:          s.Metadata.AppShortId,
		AppName:          s.Component.Name,
		InputImage:       "kennethreitz/httpbin",
		InputVersion:     "latest",
		OutputRepository: s.EcrRef.RepositoryName,
		OutputVersion:    s.EcrRef.Tag,
	}

	if err := tmpl.Execute(buf, args); err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), waypointv1.Hcl_HCL, nil
}
