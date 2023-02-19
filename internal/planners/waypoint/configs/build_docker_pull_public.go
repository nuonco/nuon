package configs

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/go-playground/validator/v10"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

const (
	publicDockerPullBuildTmplName string = "public-image-build"
)

// NewPublicDockerPullBuild returns a builder that renders our hardcoded sample application
func NewPublicDockerPullBuild(v *validator.Validate, opts ...Option) (*publicDockerPullBuild, error) {
	baseBuilder, err := newBaseBuilder(v, opts...)
	if err != nil {
		return nil, err
	}

	if baseBuilder.PublicImageSource == nil {
		return nil, fmt.Errorf("public image source not provided")
	}

	return &publicDockerPullBuild{baseBuilder}, nil
}

type publicDockerPullBuild struct {
	*baseBuilder
}

var _ Builder = (*publicDockerPullBuild)(nil)

var publicDockerPullBuildTmpl string = `
project = "{{.WaypointRef.Project}}"

app "{{.WaypointRef.App}}" {
  build {
    use "docker-pull" {
      image = "{{.PublicImageSource.Image}}"
      tag   = "{{.PublicImageSource.Tag}}"
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

func (s *publicDockerPullBuild) Render() ([]byte, waypointv1.Hcl_Format, error) {
	tmpl, err := template.New(publicDockerPullBuildTmplName).Parse(publicDockerPullBuildTmpl)
	if err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to parse static template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), waypointv1.Hcl_HCL, nil
}
