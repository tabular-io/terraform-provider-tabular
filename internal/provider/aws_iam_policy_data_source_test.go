package provider

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

const testIAMPolicyDataSourceConfig = `
data "tabular_aws_iam_policy" "test" {
	bucket = "my-bucket-name"
}
`

func TestAWSIAMPolicyValidJSON(t *testing.T) {
	assert.True(t, json.Valid([]byte(IAMReadWritePolicy("my-bucket-name"))))
	assert.True(t, json.Valid([]byte(IAMReadOnlyPolicy("my-bucket-name"))))
	assert.True(t, json.Valid([]byte(AssumeRolePolicy("some-external-id-1234"))))
}

func TestAWSIAMPolicyDataSource(t *testing.T) {
	externalId := os.Getenv("TABULAR_ORGANIZATION_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { accPreCheck(t) },
		ProtoV6ProviderFactories: accProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testIAMPolicyDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.tabular_aws_iam_policy.test", "bucket", "my-bucket-name"),
					resource.TestCheckResourceAttr("data.tabular_aws_iam_policy.test", "iam_read_write_policy", IAMReadWritePolicy("my-bucket-name")),
					resource.TestCheckResourceAttr("data.tabular_aws_iam_policy.test", "iam_read_only_policy", IAMReadOnlyPolicy("my-bucket-name")),
					resource.TestCheckResourceAttr("data.tabular_aws_iam_policy.test", "assume_role_policy", AssumeRolePolicy(externalId)),
				),
			},
		},
	})
}
