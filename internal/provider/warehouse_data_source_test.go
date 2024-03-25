package provider

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWarehouseDataSourceByName(t *testing.T) {
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	name := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWarehouseDataSourceByNameConfig(bucketName, roleArn, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tabular_warehouse.test", "name", name),
					resource.TestCheckResourceAttr("data.tabular_warehouse.test", "region", "us-west-2"),
				),
			},
		},
	})
}

func testAccWarehouseDataSourceByNameConfig(bucketName, roleArn, name string) string {
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

data "tabular_warehouse" "test" {
 name = tabular_warehouse.test.name
}
`, bucketName, roleArn, name)
}

func TestAccWarehouseDataSourceById(t *testing.T) {
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	name := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWarehouseDataSourceByIdConfig(bucketName, roleArn, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tabular_warehouse.test", "name", name),
					resource.TestCheckResourceAttr("data.tabular_warehouse.test", "region", "us-west-2"),
				),
			},
		},
	})
}

func testAccWarehouseDataSourceByIdConfig(bucketName, roleArn, name string) string {
	return fmt.Sprintf(
		`
resource "tabular_s3_storage_profile" "test" {
  region = "us-west-2"
  s3_bucket_name = "%s"
  role_arn = "%s"
}

resource "tabular_warehouse" "test" {
  name            = "%s"
  storage_profile = tabular_s3_storage_profile.test.id
}

data "tabular_warehouse" "test" {
  id = tabular_warehouse.test.id
}
`, bucketName, roleArn, name)
}
