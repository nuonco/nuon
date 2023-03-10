package configs

// this will be for the docker build
import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/go-playground/validator/v10"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

const (
	dockerBuildTmplName string = "docker-build"
)

// NewDockerBuild returns a builder that renders our hardcoded sample application
func NewDockerBuild(v *validator.Validate, opts ...Option) (*dockerBuild, error) {
	baseBuilder, err := newBaseBuilder(v, opts...)
	if err != nil {
		return nil, err
	}

	if baseBuilder.DockerCfg == nil {
		return nil, fmt.Errorf("docker config not provided")
	}

	return &dockerBuild{baseBuilder}, nil
}

type dockerBuild struct {
	*baseBuilder
}

var _ Builder = (*dockerBuild)(nil)

var dockerBuildTmpl string = `
project = "{{.WaypointRef.Project}}"

app "{{.WaypointRef.App}}" {
  build {
    use "docker" {
      dockerfile = "{{.DockerCfg.Dockerfile}}"
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

func (s *dockerBuild) Render() ([]byte, waypointv1.Hcl_Format, error) {
	tmpl, err := template.New(dockerBuildTmplName).Parse(dockerBuildTmpl)
	if err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to parse static template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), waypointv1.Hcl_HCL, nil
}
