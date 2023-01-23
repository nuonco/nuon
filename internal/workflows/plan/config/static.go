package config

import (
	"bytes"
	"fmt"
	"html/template"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
)

// NewStaticBuilder returns a builder that renders our hardcoded sample application
func NewStaticBuilder() *staticBuilder {
	return &staticBuilder{}
}

var staticBuildTmpl string = `
project = "{{.Project}}"

app "main" {
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

type buildTmplArgs struct {
	Project          string
	InputImage       string
	InputVersion     string
	OutputRepository string
	OutputVersion    string
}

type staticBuilder struct {
	ecrRef    *planv1.ECRRepositoryRef
	metadata  *planv1.Metadata
	component *componentv1.Component
}

var _ Builder = (*staticBuilder)(nil)

func (s *staticBuilder) WithMetadata(metadata *planv1.Metadata) {
	s.metadata = metadata
}
func (s *staticBuilder) WithECRRef(ecrRef *planv1.ECRRepositoryRef) {
	s.ecrRef = ecrRef
}
func (s *staticBuilder) WithComponent(component *componentv1.Component) {
	s.component = component
}
func (s *staticBuilder) Render() ([]byte, waypointv1.Hcl_Format, error) {
	tmpl, err := template.New("build-config").Parse(staticBuildTmpl)
	if err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to parse static template: %w", err)
	}

	buf := new(bytes.Buffer)
	args := buildTmplArgs{
		Project:          s.metadata.AppId,
		InputImage:       "kennethreitz/httpbin",
		InputVersion:     "latest",
		OutputRepository: s.ecrRef.RepositoryName,
		OutputVersion:    s.ecrRef.Tag,
	}

	if err := tmpl.Execute(buf, args); err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), waypointv1.Hcl_HCL, nil
}
