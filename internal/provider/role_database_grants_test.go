package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"math/rand"
	"os"
	"testing"
)

func TestAccRoleDatabaseGrantsWithoutGrants(t *testing.T) {
	testId := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	privilege := "FUTURE_DROP_TABLE"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDatabaseGrantsConfigWithoutGrants(bucketName, roleArn, testId, testId, testId, privilege),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_database_grants.test", "privileges.0", privilege),
					resource.TestCheckNoResourceAttr("tabular_role_database_grants.test", "privileges_with_grant"),
				),
			},
		},
	})
}

func testAccRoleDatabaseGrantsConfigWithoutGrants(bucketName, roleArn, warehouseName, databaseName, tabularRole, privilege string) string {
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
  warehouse_id    = tabular_warehouse.test.id
}

resource "tabular_role" "test" {
	name = "%s"
}

resource "tabular_role_database_grants" "test" {
	role_id = tabular_role.test.id
    warehouse_id = tabular_warehouse.test.id
	database_id = tabular_database.test.id
    privileges = [
		"%s"	
	]
}
	
`, bucketName, roleArn, warehouseName, databaseName, tabularRole, privilege)
}

func TestAccRoleDatabaseGrantsWithGrants(t *testing.T) {
	testId := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	privilege := "MODIFY_DATABASE"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDatabaseGrantsWithGrantsConfig(bucketName, roleArn, testId, testId, testId, privilege),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_database_grants.test", "privileges_with_grant.0", privilege),
					resource.TestCheckNoResourceAttr("tabular_role_database_grants.test", "privileges"),
				),
			},
		},
	})
}

func testAccRoleDatabaseGrantsWithGrantsConfig(bucketName, roleArn, warehouseName, databaseName, tabularRole, privilege string) string {
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
  warehouse_id    = tabular_warehouse.test.id
}

resource "tabular_role" "test" {
	name = "%s"
}

resource "tabular_role_database_grants" "test" {
	role_id = tabular_role.test.id
    warehouse_id = tabular_warehouse.test.id
	database_id = tabular_database.test.id
	privileges_with_grant = [
		"%s"
	]
}
	
`, bucketName, roleArn, warehouseName, databaseName, tabularRole, privilege)
}

func TestAccRoleDatabaseGrantsWithBothGrants(t *testing.T) {
	testId := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")
	privilege := "FUTURE_DROP_TABLE"
	privilegeWithGrant := "MODIFY_DATABASE"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDatabaseGrantsWithBothGrantsConfig(bucketName, roleArn, testId, testId, testId, privilege, privilegeWithGrant),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_database_grants.test", "privileges.0", privilege),
					resource.TestCheckResourceAttr("tabular_role_database_grants.test", "privileges_with_grant.0", privilegeWithGrant),
				),
			},
		},
	})
}

func testAccRoleDatabaseGrantsWithBothGrantsConfig(bucketName, roleArn, warehouseName, databaseName, tabularRole, privilege, privilegeWithGrant string) string {
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
  warehouse_id    = tabular_warehouse.test.id
}

resource "tabular_role" "test" {
	name = "%s"
}

resource "tabular_role_database_grants" "test" {
	role_id = tabular_role.test.id
    warehouse_id = tabular_warehouse.test.id
	database_id = tabular_database.test.id
  	privileges   = [
    	"%s",
	]
	privileges_with_grant = [
		"%s"
	]
}
	
`, bucketName, roleArn, warehouseName, databaseName, tabularRole, privilege, privilegeWithGrant)
}

func TestAccRoleDatabaseMultipleGrants(t *testing.T) {
	testId := fmt.Sprintf("tf-acc-test-%d", rand.Intn(100))
	bucketName := os.Getenv("TABULAR_AWS_S3_BUCKET")
	roleArn := os.Getenv("TABULAR_AWS_IAM_ROLE_ARN")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDatabaseMultipleGrantsConfig(bucketName, roleArn, testId, testId, testId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("tabular_role_database_grants.test", "privileges.1", "MODIFY_DATABASE"),
					resource.TestCheckResourceAttr("tabular_role_database_grants.test", "privileges.0", "CREATE_TABLE"),
					resource.TestCheckResourceAttr("tabular_role_database_grants.test", "privileges_with_grant.0", "LIST_TABLES"),
				),
			},
		},
	})
}

func testAccRoleDatabaseMultipleGrantsConfig(bucketName, roleArn, warehouseName, databaseName, tabularRole string) string {
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
  warehouse_id    = tabular_warehouse.test.id
}

resource "tabular_role" "test" {
	name = "%s"
}

resource "tabular_role_database_grants" "test" {
	role_id = tabular_role.test.id
    warehouse_id = tabular_warehouse.test.id
	database_id = tabular_database.test.id

  	privileges   = [
    	"MODIFY_DATABASE",
		"CREATE_TABLE"
	]
	privileges_with_grant = [
		"LIST_TABLES"
	]
}
	
`, bucketName, roleArn, warehouseName, databaseName, tabularRole)
}
