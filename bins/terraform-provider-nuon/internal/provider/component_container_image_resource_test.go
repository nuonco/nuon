package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccComponentContainerImageResource(args ContainerImageComponentResourceModel) string {
	return fmt.Sprintf(providerConfig+`
resource "nuon_app" "my_app" {
    name = "My App"
}

resource "nuon_container_image_component" "my_component" {
    app_id = nuon_app.my_app.id
    name = "%s"
    sync_only = %v

    public = {
        image_url = "%s"
        tag = "%s"
    }

}
`,
		args.Name.ValueString(),
		args.SyncOnly.ValueBool(),
		args.Public.ImageURL.ValueString(),
		args.Public.Tag.ValueString(),
	)
}

func TestComponentContainerImageResource(t *testing.T) {
	createArgs := ContainerImageComponentResourceModel{
		Name:        types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
		SyncOnly:    types.BoolValue(true),
		BasicDeploy: nil,
		AwsEcr:      nil,
		Public: &Public{
			ImageURL: types.StringValue("kennethreitz/httpbin"),
			Tag:      types.StringValue("latest"),
		},
		EnvVar: []EnvVar{},
	}

	updateArgs := ContainerImageComponentResourceModel{
		Name:        types.StringValue(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
		SyncOnly:    createArgs.SyncOnly,
		BasicDeploy: createArgs.BasicDeploy,
		AwsEcr:      createArgs.AwsEcr,
		Public:      createArgs.Public,
		EnvVar:      createArgs.EnvVar,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccComponentContainerImageResource(createArgs),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_container_image_component.my_component", "name", createArgs.Name.ValueString()),
					resource.TestCheckResourceAttr("nuon_container_image_component.my_component", "sync_only", createArgs.SyncOnly.String()),
					resource.TestCheckResourceAttr("nuon_container_image_component.my_component", "public.image_url", createArgs.Public.ImageURL.ValueString()),
					resource.TestCheckResourceAttr("nuon_container_image_component.my_component", "public.tag", createArgs.Public.Tag.ValueString()),
				),
			},
			// Import State
			{
				ResourceName:      "nuon_container_image_component.my_component",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: testAccComponentContainerImageResource(updateArgs),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_container_image_component.my_component", "name", updateArgs.Name.ValueString()),
					resource.TestCheckResourceAttr("nuon_container_image_component.my_component", "sync_only", updateArgs.SyncOnly.String()),
					resource.TestCheckResourceAttr("nuon_container_image_component.my_component", "public.image_url", updateArgs.Public.ImageURL.ValueString()),
					resource.TestCheckResourceAttr("nuon_container_image_component.my_component", "public.tag", updateArgs.Public.Tag.ValueString()),
				),
			},
			// Delete testing will happen automatically.
		},
	})
}
