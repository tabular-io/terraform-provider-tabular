package provider

import (
	"fmt"
	"golang.org/x/exp/rand"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWarehouse(t *testing.T) {
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	name := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWarehouseConfig(bucketName, roleArn, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_warehouse.test", "name", name),
				),
			},
		},
	})
}

func testAccWarehouseConfig(bucketName, roleArn, name string) string {
	return fmt.Sprintf(`
resource "tabular_s3_storage_profile" "test" {
  region = "us-west-2"
  s3_bucket_name = "%s"
  role_arn = "%s"
}

resource "tabular_warehouse" "test" {
  name            = "%s"
  storage_profile = tabular_s3_storage_profile.test.id
}
`, bucketName, roleArn, name)
}
