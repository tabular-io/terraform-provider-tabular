package data_sources

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RoleDataSource{}
var _ datasource.DataSourceWithConfigure = &RoleDataSource{}

func NewRoleDataSource() datasource.DataSource {
	return &RoleDataSource{}
}

// RoleDataSource defines the data source implementation.
type RoleDataSource struct {
	client *tabular.Client
}

// RoleDataSourceModel describes the data source data model.
type RoleDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *RoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *RoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Role data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Role Name",
				Required:            true,
			},
		},
	}
}

func (d *RoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tabular.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *RoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.client.GetRole(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed fetching role", err.Error())
		return
	}
	if role == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Id = types.StringValue(role.Id)
	data.Name = types.StringValue(role.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
