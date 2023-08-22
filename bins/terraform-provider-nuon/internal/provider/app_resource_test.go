package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccAppResource(name string) string {
	return fmt.Sprintf(providerConfig+`
resource "nuon_app" "my_app" {
    name = "%s"
}
`,
		name,
	)
}

func TestAppResource(t *testing.T) {
	appName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	updatedName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccAppResource(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_app.my_app", "name", appName),
				),
			},
			// ImportState
			{
				ResourceName:      "nuon_app.my_app",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: testAccAppResource(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nuon_app.my_app", "name", updatedName),
				),
			},
			// Delete testing will happen automatically.
		},
	})
}
