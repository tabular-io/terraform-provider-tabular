package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWarehouseDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWarehouseDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tabular_warehouse.test", "name", "sandbox"),
					resource.TestCheckResourceAttr("data.tabular_warehouse.test", "region", "us-west-2"),
				),
			},
		},
	})
}

const testAccWarehouseDataSourceConfig = `
data "tabular_warehouse" "test" {
  name = "sandbox"
}
`
