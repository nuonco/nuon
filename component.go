package waypoint

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const (
	HCLBuild int = iota
	HCLDeploy
)

// Component represents a component from the api. It only supports public container images now, but will support helm,
// terraform and more later.
type Component struct {
	Type              string `json:"type" validate:"required"`
	ContainerImageURL string `json:"container_image_url" validate:"required"`
	Name              string
	ID                string
}

// GenerateHCL will create a waypoint hclwrite.File which can be saved to file
// in s3
func (c *Component) GenerateHCL(t int) *hclwrite.File {
	f := hclwrite.NewEmptyFile()
	mainBody := f.Body()

	// start building app "component-name"{} from name of component
	appBlock := mainBody.AppendNewBlock("app", []string{c.Name})
	appBody := appBlock.Body()

	if t == HCLBuild {
		/*
			 build {
			   use "docker" {
					 registry {
						 image: "component.ContainerImageURL"
						 tag: "latest"
					 }
				 }

			   registry {
			     # ...
			   }

				  hook {
				    # ...
				  }
				}
		*/
		buildBlock := appBody.AppendNewBlock("build", nil)
		buildBody := buildBlock.Body()
		dockerBlock := buildBody.AppendNewBlock("use", []string{"docker-pull"})
		dockerBody := dockerBlock.Body()
		registryBlock := dockerBody.AppendNewBlock("registry", nil)
		registryBody := registryBlock.Body()
		registryBody.SetAttributeValue("image", cty.StringVal(c.ContainerImageURL))
		registryBody.SetAttributeValue("tag", cty.StringVal("latest"))
		mainBody.AppendNewline()
		deployBlock := appBody.AppendNewBlock("deploy", nil)
		deployBody := deployBlock.Body()
		_ = deployBody.AppendNewBlock("use", []string{"kubernetes"})
	}

	if t == HCLDeploy {
		/*
			deploy {
			    use "kubernetes" {}

				   hook {
				     # ...
				   }
				 }

		*/
		mainBody.AppendNewline()
		deployBlock := appBody.AppendNewBlock("deploy", nil)
		deployBody := deployBlock.Body()
		_ = deployBody.AppendNewBlock("use", []string{"kubernetes"})
	}
	return f
}
