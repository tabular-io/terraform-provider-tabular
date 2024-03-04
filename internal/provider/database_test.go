package provider

import (
	"fmt"
	"golang.org/x/exp/rand"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatabase(t *testing.T) {
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	name := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig(bucketName, roleArn, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_database.test", "name", name),
				),
			},
		},
	})
}

func testAccDatabaseConfig(bucketName, roleArn, name string) string {
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

resource "tabular_database" "test" {
  name            = "%s"
  warehouse_id = tabular_warehouse.test.id
}
`, bucketName, roleArn, name, name)
}
