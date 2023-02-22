package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
	"os"
)

var defaultEndpoint = "https://api.tabular.io"
var defaultTokenEndpoint = "https://api.tabular.io/ws/v1/oauth/tokens"

var _ provider.Provider = &TabularProvider{}

type TabularProvider struct {
}

func (p *TabularProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tabular"
}

func (p *TabularProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	_, credentialSet := os.LookupEnv("TABULAR_CREDENTIAL")
	resp.Schema = schema.Schema{
		Description: "",
		Attributes: map[string]schema.Attribute{
			"token_endpoint": schema.StringAttribute{
				Description: "Endpoint for authentication. May also be provided via TABULAR_TOKEN_ENDPOINT environment variable.",
				Optional:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "Endpoint for Tabular API. May also be provided via TABULAR_ENDPOINT environment variable.",
				Optional:    true,
			},
			"credential": schema.StringAttribute{
				Description: "Tabular Credential. May also be provided via TABULAR_CREDENTIAL environment variable.",
				Required:    !credentialSet,
				Optional:    credentialSet,
				Sensitive:   true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Tabular Organization ID. May also be provided via TABULAR_ORGANIZATION_ID environment variable.",
				Required:    true,
			},
		},
	}
}

type TabularProviderModel struct {
	TokenEndpoint  types.String `tfsdk:"token_endpoint"`
	Endpoint       types.String `tfsdk:"endpoint"`
	Credential     types.String `tfsdk:"credential"`
	OrganizationId types.String `tfsdk:"organization_id"`
}

func ensureProviderConfigOption(
	attr types.String,
	attrName string,
	envVar string,
	defaultValue *string,
) (*string, error) {
	if attr.IsUnknown() {
		return nil, fmt.Errorf("%s depends on values that cannot be known until apply time", attrName)
	} else if attr.ValueString() == "" {
		value, valueSet := os.LookupEnv(envVar)
		if !valueSet {
			if defaultValue != nil {
				return defaultValue, nil
			} else {
				return nil, fmt.Errorf(
					"%s must have a value. Either set %s in provider config or set the %s enviornment variable",
					attrName, attrName, envVar,
				)
			}
		} else {
			return &value, nil
		}
	} else {
		value := attr.ValueString()
		return &value, nil
	}
}

func (p *TabularProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config TabularProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, err := ensureProviderConfigOption(
		config.Endpoint,
		"endpoint",
		"TABULAR_ENDPOINT",
		&defaultEndpoint,
	)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("endpoint"), "Endpoint Invalid", err.Error())
	}

	tokenEndpoint, err := ensureProviderConfigOption(
		config.TokenEndpoint,
		"token_endpoint",
		"TABULAR_TOKEN_ENDPOINT",
		&defaultTokenEndpoint,
	)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("token_endpoint"), "Token Endpoint Invalid", err.Error())
	}

	orgId, err := ensureProviderConfigOption(
		config.OrganizationId,
		"organization_id",
		"TABULAR_ORGANIZATION_ID",
		nil,
	)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("organization_id"), "Organization ID Invalid", err.Error())
	}

	credential, err := ensureProviderConfigOption(
		config.Credential,
		"credential",
		"TABULAR_CREDENTIAL",
		nil,
	)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("credential"), "Credential Invalid", err.Error())
	}

	if resp.Diagnostics.HasError() {
		return
	}
	client, err := tabular.NewClient(
		*endpoint,
		*tokenEndpoint,
		*orgId,
		*credential,
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed setting up Tabular Client", err.Error())
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TabularProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRoleResource,
	}
}

func (p *TabularProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRoleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TabularProvider{}
	}
}
