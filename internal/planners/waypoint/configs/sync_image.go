package configs

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/go-playground/validator/v10"
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

const (
	syncImageTmplName string = "sync-image"
)

// Note(jm): essentially, we need to do this https://docs.aws.amazon.com/AmazonECR/latest/userguide/registry_auth.html
type SyncImageSource struct {
	// this is the output of aws ecr get-login-password --region region | docker login --username AWS
	// --password-stdin aws_account_id.dkr.ecr.region.amazonaws.com
	RegistryToken string `validate:"required"`

	// this should be the registry uri: aws_account_id.dkr.ecr.region.amazonaws.com
	ServerAddress string `validate:"required"`

	// this should be AWS
	Username string `validate:"required"`

	Image string `validate:"required"`
	Tag   string `validate:"required"`
}

// NewSyncImageBuilder returns a builder which will render a job to sync an image into an end waypoint runner's location
func NewSyncImageBuilder(v *validator.Validate, opts ...baseBuilderOption) (*syncImageBuilder, error) {
	baseBuilder, err := newBaseBuilder(v, opts...)
	if err != nil {
		return nil, err
	}
	return &syncImageBuilder{baseBuilder}, nil
}

var _ Builder = (*httpbinBuildBuilder)(nil)

type syncImageBuilder struct {
	*baseBuilder
}

var syncImageTmpl string = `
project = "{{.WaypointRef.Project}}"

app "{{.WaypointRef.AppName}}" {
  build {
    use "docker-pull" {
      image = "{{.SyncImageSource.Tag}}"
      tag   = "{{.SyncImageSource.Tag}}"

      auth {
	registryToken = "{{.SyncImageSource.RegistryToken}}"
	serverAddress = "{{.SyncImageSource.ServerAddress}}"
      }
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

func (s *syncImageBuilder) Render() ([]byte, waypointv1.Hcl_Format, error) {
	tmpl, err := template.New(syncImageTmplName).Parse(syncImageTmpl)
	if err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to parse %s template: %w", syncImageTmplName, err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, waypointv1.Hcl_HCL, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), waypointv1.Hcl_HCL, nil
}
