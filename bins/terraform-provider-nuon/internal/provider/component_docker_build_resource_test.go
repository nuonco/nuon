package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccComponentDockerBuildResource(args DockerBuildComponentResourceModel) string {
	return fmt.Sprintf(providerConfig+`
resource "nuon_app" "my_app" {
    name = "My App"
}

resource "nuon_docker_build_component" "my_component" {
    app_id = nuon_app.my_app.id
    name = %s
    sync_only = %v
    dockerfile = %s

    public_repo = {
        repo = %s
        directory = %s
        branch = %s
    }
}
`,
		args.Name,
		args.SyncOnly,
		args.Dockerfile,
		args.PublicRepo.Repo,
		args.PublicRepo.Directory,
		args.PublicRepo.Branch,
	)
}

func TestComponentDockerBuildResource(t *testing.T) {
	createArgs := DockerBuildComponentResourceModel{
		Name:       types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
		SyncOnly:   types.BoolValue(true),
		Dockerfile: types.StringValue("Dockerfile"),
		PublicRepo: &PublicRepo{
			Repo:      types.StringValue("https://github.com/postmanlabs/httpbin.git"),
			Directory: types.StringValue("."),
			Branch:    types.StringValue("master"),
		},
		ConnectedRepo: nil,
		BasicDeploy:   nil,
		EnvVar:        []EnvVar{},
	}

	updateArgs := DockerBuildComponentResourceModel{
		Name:       types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
		SyncOnly:   types.BoolValue(true),
		Dockerfile: types.StringValue("Dockerfile"),
		PublicRepo: &PublicRepo{
			Repo:      types.StringValue("https://github.com/postmanlabs/httpbin.git"),
			Directory: types.StringValue("."),
			Branch:    types.StringValue("master"),
		},
		ConnectedRepo: nil,
		BasicDeploy:   nil,
		EnvVar:        []EnvVar{},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccComponentDockerBuildResource(createArgs),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "name", createArgs.Name.ValueString()),
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "sync_only", createArgs.SyncOnly.String()),
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "public_repo.repo", createArgs.PublicRepo.Repo.ValueString()),
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "public_repo.directory", createArgs.PublicRepo.Directory.ValueString()),
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "public_repo.branch", createArgs.PublicRepo.Branch.ValueString()),
				),
			},
			// Import State
			{
				ResourceName:      "nuon_docker_build_component.my_component",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: testAccComponentDockerBuildResource(updateArgs),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "name", updateArgs.Name.ValueString()),
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "sync_only", updateArgs.SyncOnly.String()),
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "public_repo.repo", updateArgs.PublicRepo.Repo.ValueString()),
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "public_repo.directory", updateArgs.PublicRepo.Directory.ValueString()),
					resource.TestCheckResourceAttr("nuon_docker_build_component.my_component", "public_repo.branch", updateArgs.PublicRepo.Branch.ValueString()),
				),
			},
			// Delete testing will happen automatically.
		},
	})
}
