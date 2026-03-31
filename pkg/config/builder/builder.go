package builder

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
)

// AttributeHandler applies an attribute's configuration to an AppConfig.
type AttributeHandler func(cfg *config.AppConfig)

// Builder builds a config.AppConfig from user-selected app attributes.
type Builder interface {
	Build(appAttributes []string) (*config.AppConfig, error)
}

// New returns a Builder for the given cloud provider.
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

// Attribute constants matching the UI selections.
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

// Default CloudFormation nested template URLs for AWS stacks.
const (
	defaultAWSVPCTemplateURL    = "https://nuon-artifacts.s3.us-west-2.amazonaws.com/aws-cloudformation-templates/v0.2.1/vpc/eks/default/stack.yaml"
	defaultAWSRunnerTemplateURL = "https://nuon-artifacts.s3.us-west-2.amazonaws.com/aws-cloudformation-templates/v0.2.1/runner/asg/stack.yaml"
)
