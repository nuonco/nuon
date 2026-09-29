package builder

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
)

type AttributeHandler func(cfg *config.AppConfig)

type Builder interface {
	Build(appAttributes []string) (*config.AppConfig, error)
}

func New(cloudProvider string) (Builder, error) {
	switch cloudProvider {
	case "aws":
		return newAWSBuilder(), nil
	case "gcp":
		return newGCPBuilder(), nil
	case "azure":
		return newAzureBuilder(), nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", cloudProvider)
	}
}

const (
	AttributeTerraform     = "terraform"
	AttributeHelmCharts    = "helm_charts"
	AttributeKubernetes    = "kubernetes"
	AttributeLambda        = "lambda"
	AttributeDockerImage   = "docker_image"
	AttributeCustomScripts = "custom_scripts"
)

const sampleRepo = "nuonco/example-app-configs"
const sampleBranch = "main"

const (
	defaultAWSVPCTemplateURL    = "https://nuon-artifacts.s3.us-west-2.amazonaws.com/aws-cloudformation-templates/v0.4.0/vpc/eks/default/stack.yaml"
	defaultAWSRunnerTemplateURL = "https://nuon-artifacts.s3.us-west-2.amazonaws.com/aws-cloudformation-templates/v0.4.0/runner/asg/stack.yaml"
)
