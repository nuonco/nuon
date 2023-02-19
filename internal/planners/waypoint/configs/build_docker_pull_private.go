package configs

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/go-playground/validator/v10"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

const (
	privateDockerPullBuildTmplName string = "private-image-build"
)

// NewPrivateDockerPullBuild returns a builder that renders our hardcoded sample application
func NewPrivateDockerPullBuild(v *validator.Validate, opts ...Option) (*privateDockerPullBuild, error) {
	baseBuilder, err := newBaseBuilder(v, opts...)
	if err != nil {
		return nil, err
	}

	if baseBuilder.PrivateImageSource == nil {
		return nil, fmt.Errorf("private image source not provided")
	}

	return &privateDockerPullBuild{baseBuilder}, nil
}

type privateDockerPullBuild struct {
	*baseBuilder
}

var _ Builder = (*privateDockerPullBuild)(nil)

var privateDockerPullBuildTmpl string = `
project = "{{.WaypointRef.Project}}"

app "{{.WaypointRef.App}}" {
  build {
    use "docker-pull" {
      image = "{{.PrivateImageSource.Image}}"
      tag   = "{{.PrivateImageSource.Tag}}"

      encoded_auth = base64encode(
	jsonencode({
	  username = "{{.PrivateImageSource.Username}}",
	  password = "{{.PrivateImageSource.RegistryToken}}"
	})
      )
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

func (s *privateDockerPullBuild) Render() ([]byte, waypointv1.Hcl_Format, error) {
	tmpl, err := template.New(privateDockerPullBuildTmplName).Parse(privateDockerPullBuildTmpl)
	if err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to parse static template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), waypointv1.Hcl_HCL, nil
}
