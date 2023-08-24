package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccComponentTerraformModuleResource(args TerraformModuleComponentResourceModel) string {
	return fmt.Sprintf(providerConfig+`
resource "nuon_app" "my_app" {
    name = "My App"
}

resource "nuon_terraform_module_component" "my_component" {
    app_id = nuon_app.my_app.id
    name = %s

    public_repo = {
        repo = %s
        branch = %s
        directory = %s
    }
}
`,
		args.Name,
		args.PublicRepo.Repo,
		args.PublicRepo.Branch,
		args.PublicRepo.Directory,
	)
}

func TestComponentTerraformModuleResource(t *testing.T) {
	createArgs := TerraformModuleComponentResourceModel{
		Name: types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
		PublicRepo: &PublicRepo{
			Repo:      types.StringValue("my-github-org/" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
			Branch:    types.StringValue("foobar"),
			Directory: types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
		},
	}

	updateArgs := TerraformModuleComponentResourceModel{
		Name: types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
		PublicRepo: &PublicRepo{
			Repo:      types.StringValue("my-github-org/" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
			Branch:    types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
			Directory: types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccComponentTerraformModuleResource(createArgs),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_terraform_module_component.my_component", "name", createArgs.Name.ValueString()),
					resource.TestCheckResourceAttr("nuon_terraform_module_component.my_component", "public_repo.repo", createArgs.PublicRepo.Repo.ValueString()),
					resource.TestCheckResourceAttr("nuon_terraform_module_component.my_component", "public_repo.branch", createArgs.PublicRepo.Branch.ValueString()),
					resource.TestCheckResourceAttr("nuon_terraform_module_component.my_component", "public_repo.directory", createArgs.PublicRepo.Directory.ValueString()),
				),
			},
			// Import State
			{
				ResourceName:      "nuon_terraform_module_component.my_component",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: testAccComponentTerraformModuleResource(updateArgs),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_terraform_module_component.my_component", "name", updateArgs.Name.ValueString()),
					resource.TestCheckResourceAttr("nuon_terraform_module_component.my_component", "public_repo.repo", updateArgs.PublicRepo.Repo.ValueString()),
					resource.TestCheckResourceAttr("nuon_terraform_module_component.my_component", "public_repo.branch", updateArgs.PublicRepo.Branch.ValueString()),
					resource.TestCheckResourceAttr("nuon_terraform_module_component.my_component", "public_repo.directory", updateArgs.PublicRepo.Directory.ValueString()),
				),
			},
			// Delete testing will happen automatically.
		},
	})
}
