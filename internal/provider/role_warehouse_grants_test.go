package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"math/rand"
	"os"
	"testing"
)

func TestAccWarehouseRoleGrantsWithoutGrants(t *testing.T) {
	testId := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	privilege := "FUTURE_DROP_TABLE"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWarehouseRoleGrantsConfigWithoutGrants(bucketName, roleArn, testId, testId, privilege),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_warehouse_grants.test", "privileges.0", privilege),
					resource.TestCheckNoResourceAttr("tabular_role_warehouse_grants.test", "privileges_with_grant"),
				),
			},
		},
	})
}

func testAccWarehouseRoleGrantsConfigWithoutGrants(bucketName, roleArn, name, tabularRole, privilege string) string {
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

resource "tabular_role" "test" {
	name = "%s"
}

resource "tabular_role_warehouse_grants" "test" {
	role_id = tabular_role.test.id
    warehouse_id = tabular_warehouse.test.id
    privileges = [
		"%s"	
	]
}
	
`, bucketName, roleArn, name, tabularRole, privilege)
}

func TestAccWarehouseRoleGrantsWithGrants(t *testing.T) {
	testId := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	privilege := "FUTURE_MODIFY_DATABASE"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWarehouseRoleGrantsWithGrantsConfig(bucketName, roleArn, testId, testId, privilege),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_warehouse_grants.test", "privileges_with_grant.0", privilege),
					resource.TestCheckNoResourceAttr("tabular_role_warehouse_grants.test", "privileges"),
				),
			},
		},
	})
}

func testAccWarehouseRoleGrantsWithGrantsConfig(bucketName, roleArn, name, tabularRole, privilege string) string {
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

resource "tabular_role" "test" {
	name = "%s"
}

resource "tabular_role_warehouse_grants" "test" {
	role_id = tabular_role.test.id
    warehouse_id = tabular_warehouse.test.id
	privileges_with_grant = [
		"%s"
	]
}
	
`, bucketName, roleArn, name, tabularRole, privilege)
}

func TestAccWarehouseRoleGrantsWithBothGrants(t *testing.T) {
	testId := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	privilege := "FUTURE_DROP_TABLE"
	privilegeWithGrant := "FUTURE_MODIFY_DATABASE"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWarehouseRoleGrantsWithBothGrantsConfig(bucketName, roleArn, testId, testId, privilege, privilegeWithGrant),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_warehouse_grants.test", "privileges.0", privilege),
					resource.TestCheckResourceAttr("tabular_role_warehouse_grants.test", "privileges_with_grant.0", privilegeWithGrant),
				),
			},
		},
	})
}

func testAccWarehouseRoleGrantsWithBothGrantsConfig(bucketName, roleArn, name, tabularRole, privilege, privilegeWithGrant string) string {
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

resource "tabular_role" "test" {
	name = "%s"
}

resource "tabular_role_warehouse_grants" "test" {
	role_id = tabular_role.test.id
    warehouse_id = tabular_warehouse.test.id
  	privileges   = [
    	"%s",
	]
	privileges_with_grant = [
		"%s"
	]
}
	
`, bucketName, roleArn, name, tabularRole, privilege, privilegeWithGrant)
}

func TestAccWarehouseRoleMultipleGrants(t *testing.T) {
	testId := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWarehouseRoleMultipleGrantsConfig(bucketName, roleArn, testId, testId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_warehouse_grants.test", "privileges.1", "MODIFY_WAREHOUSE"),
					resource.TestCheckResourceAttr("tabular_role_warehouse_grants.test", "privileges.0", "FUTURE_LIST_TABLES"),
					resource.TestCheckResourceAttr("tabular_role_warehouse_grants.test", "privileges_with_grant.0", "FUTURE_MODIFY_DATABASE"),
				),
			},
		},
	})
}

func testAccWarehouseRoleMultipleGrantsConfig(bucketName, roleArn, name, tabularRole string) string {
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

resource "tabular_role" "test" {
	name = "%s"
}

resource "tabular_role_warehouse_grants" "test" {
	role_id = tabular_role.test.id
    warehouse_id = tabular_warehouse.test.id
  	privileges   = [
    	"MODIFY_WAREHOUSE",
		"FUTURE_LIST_TABLES"
	]
	privileges_with_grant = [
		"FUTURE_MODIFY_DATABASE"
	]
}
	
`, bucketName, roleArn, name, tabularRole)
}
