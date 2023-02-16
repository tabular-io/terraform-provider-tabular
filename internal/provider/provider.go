package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
	"net/http"
	"os"
)

var _ provider.Provider = &TabularProvider{}

type TabularProvider struct {
}

func (p *TabularProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tabular"
}

func (p *TabularProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	_, credentialSet := os.LookupEnv("TABULAR_CREDENTIAL")
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Tabular Endpoint",
				Optional:            true,
			},
			"credential": schema.StringAttribute{
				MarkdownDescription: "Tabular Credential",
				Required:            !credentialSet,
				Optional:            credentialSet,
				Sensitive:           true,
			},
		},
	}
}

type TabularProviderModel struct {
	Endpoint   types.String `tfsdk:"endpoint"`
	Credential types.String `tfsdk:"credential"`
}

type ConfigStuff struct {
	Client     *http.Client
	Credential string
}

func (p *TabularProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	credential := os.Getenv("TABULAR_CREDENTIAL")

	var config TabularProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Endpoint.IsUnknown() || config.Endpoint.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("endpoint"), "Endpoint Unknown", "Can't use unknown endpoint")
	}
	if config.Credential.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("credential"), "Credential Unknown", "Can't use unknown credential")
	}

	if config.Credential.ValueString() != "" {
		credential = config.Credential.ValueString()
	}

	client, err := tabular.NewClient(config.Endpoint.ValueString(), credential)
	if err != nil {
		resp.Diagnostics.AddError("Setup", err.Error())
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TabularProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
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
