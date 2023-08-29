package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"os"
	"testing"
)

// accProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var accProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"tabular": providerserver.NewProtocol6WithError(New("test")()),
}

func accPreCheck(t *testing.T) {
	envVars := []string{
		"TABULAR_ENDPOINT",
		"TABULAR_CREDENTIAL",
		"TABULAR_TOKEN_ENDPOINT",
		"TABULAR_ORGANIZATION_ID",
		"TABULAR_AWS_S3_BUCKET",
		"TABULAR_AWS_IAM_ROLE_ARN",
	}
	for _, e := range envVars {
		value := os.Getenv(e)
		if value == "" {
			t.Fatalf("Missing required environment variable %s", e)
		}
	}
}
