package provider

import (
	"fmt"
	"golang.org/x/exp/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleMappingAWS(t *testing.T) {
	name := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	iamRoleArn := fmt.Sprintf("arn:aws:iam::123456789012:role/%s", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMappingAWSConfig(name, iamRoleArn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_mapping_aws.default", "name", name),
					resource.TestCheckResourceAttr("tabular_role_mapping_aws.default", "aws_role_arn", iamRoleArn),
					resource.TestCheckResourceAttrSet("tabular_role_mapping_aws.default", "id"),
				),
			},
		},
	})
}

func testAccRoleMappingAWSConfig(name, iamRoleArn string) string {
	return fmt.Sprintf(`
resource "tabular_role" "default" {
  name = "%s"
}

resource "tabular_role_mapping_aws" "default" {
  name    = "%s"
  role_id = tabular_role.default.id
  aws_role_arn = "%s"
}
`, name, name, iamRoleArn)
}
