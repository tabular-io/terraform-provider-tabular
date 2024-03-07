package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAccS3StorageProfileDataSource(t *testing.T) {
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	region := "us-west-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccS3StorageProfileDataSourceConfig(bucketName, roleArn, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tabular_s3_storage_profile.test", "name", bucketName),
					resource.TestCheckResourceAttr("data.tabular_s3_storage_profile.test", "role_arn", roleArn),
					resource.TestCheckResourceAttr("data.tabular_s3_storage_profile.test", "region", region),
					resource.TestCheckResourceAttrSet("data.tabular_s3_storage_profile.test", "id"),
					resource.TestCheckResourceAttrSet("data.tabular_s3_storage_profile.test", "external_id"),
					resource.TestCheckResourceAttrSet("data.tabular_s3_storage_profile.test", "organization_id"),
					resource.TestCheckResourceAttrSet("data.tabular_s3_storage_profile.test", "account_id"),
				),
			},
		},
	})
}

func testAccS3StorageProfileDataSourceConfig(bucketName, roleArn, region string) string {
	return fmt.Sprintf(
		`
resource "tabular_s3_storage_profile" "test" {
  region = "%s"
  s3_bucket_name = "%s"
  role_arn = "%s"
}

data "tabular_s3_storage_profile" "test" {
  name = "%s"

  depends_on = [tabular_s3_storage_profile.test]
}
`, region, bucketName, roleArn, bucketName)
}
