package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAppDataSource(t *testing.T) {
	t.Skip()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "nuon_app" "my_app" {
                    id = "app123"
                }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.nuon_app.my_app", "name", "My App"),
					resource.TestCheckResourceAttr("data.nuon_app.my_app", "id", "app123"),
				),
			},
		},
	})
}
