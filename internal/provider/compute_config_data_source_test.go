package provider

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

func testAccComputeConfigDataSourceConfig(bucketName, roleArn, name string) string {
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

		data "tabular_compute_config" "test" {
  			warehouse_id = tabular_warehouse.test.id
		}
	`, bucketName, roleArn, name)
}

func TestSparkConfigValidJSON(t *testing.T) {
	assert.True(t, json.Valid([]byte(GetIAMRoleMappingSparkConfig("my-bucket-name", "us-west-2"))))
}

func TestAccComputeConfigDataSource(t *testing.T) {
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	name := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeConfigDataSourceConfig(bucketName, roleArn, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tabular_compute_config.test", "spark_config", GetIAMRoleMappingSparkConfig(name, "us-west-2")),
				),
			},
		},
	})
}
