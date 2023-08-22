package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccInstallResource(args ...string) string {
	return fmt.Sprintf(providerConfig+`
resource "nuon_app" "my_app" {
    name = "My App"
}

resource "nuon_install" "my_install" {
    app_id = nuon_app.my_app.id
    name = "%s"
    region = "%s"
    iam_role_arn = "%s"
}
`,
		args[0],
		args[1],
		args[2],
	)
}

func TestInstallResource(t *testing.T) {
	createArgs := []string{
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		"us-west-2",
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
	}

	updateArgs := []string{
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		createArgs[1],
		createArgs[2],
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccInstallResource(createArgs...),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_install.my_install", "name", createArgs[0]),
					resource.TestCheckResourceAttr("nuon_install.my_install", "region", createArgs[1]),
					resource.TestCheckResourceAttr("nuon_install.my_install", "iam_role_arn", createArgs[2]),
				),
			},
			// Import State
			{
				ResourceName:      "nuon_install.my_install",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: testAccInstallResource(updateArgs...),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_install.my_install", "name", updateArgs[0]),
					resource.TestCheckResourceAttr("nuon_install.my_install", "region", updateArgs[1]),
					resource.TestCheckResourceAttr("nuon_install.my_install", "iam_role_arn", updateArgs[2]),
				),
			},
			// Delete testing will happen automatically.
		},
	})
}
