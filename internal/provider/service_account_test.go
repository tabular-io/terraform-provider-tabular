package provider

import (
	"fmt"
	"golang.org/x/exp/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceAccount(t *testing.T) {
	name := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountConfig(name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_service_account.default", "name", name),
					resource.TestCheckResourceAttrSet("tabular_service_account.default", "credential_key"),
					resource.TestCheckResourceAttrSet("tabular_service_account.default", "credential_secret"),
				),
			},
		},
	})
}

func testAccServiceAccountConfig(roleName, name string) string {
	return fmt.Sprintf(`
resource "tabular_role" "default" {
  name = "%s"
}

resource "tabular_service_account" "default" {
  name    = "%s"
  role_id = tabular_role.default.id
}

`, roleName, name)
}
