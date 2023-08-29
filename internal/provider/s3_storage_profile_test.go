package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccS3StorageProfile(t *testing.T) {
	externalId := os.Getenv("TABULAR_ORGANIZATION_ID")
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccS3StorageProfileConfig(bucketName, roleArn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_s3_storage_profile.test", "s3_bucket_name", bucketName),
					resource.TestCheckResourceAttr("tabular_s3_storage_profile.test", "region", "us-west-2"),
					resource.TestCheckResourceAttr("tabular_s3_storage_profile.test", "role_arn", roleArn),
					resource.TestCheckResourceAttr("tabular_s3_storage_profile.test", "external_id", externalId),
				),
			},
		},
	})
}

func testAccS3StorageProfileConfig(bucketName, roleArn string) string {
	return fmt.Sprintf(`
resource "tabular_s3_storage_profile" "test" {
  region = "us-west-2"
  s3_bucket_name = "%s"
  role_arn = "%s"
}
`, bucketName, roleArn)
}
