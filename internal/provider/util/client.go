package util

import (
	tabularv2 "github.com/tabular-io/tabular-sdk-go/tabular"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
)

type Client struct {
	V1             *tabular.Client
	V2             *tabularv2.APIClient
	OrganizationId *string
}
