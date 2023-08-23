package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
)

var _ datasource.DataSource = &WarehouseDataSource{}
var _ datasource.DataSourceWithConfigure = &WarehouseDataSource{}

func NewWarehouseDataSource() datasource.DataSource {
	return &WarehouseDataSource{}
}

type WarehouseDataSource struct {
	client *util.Client
}

type WarehouseDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationId types.String `tfsdk:"organization_id"`
	//Properties     types.Map    `tfsdk:"properties"`
	StorageProfile types.String `tfsdk:"storage_profile"`
}

func (d *WarehouseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_warehouse"
}

func (d *WarehouseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Warehouse data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Warehouse ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Role Name",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Computed:            true,
			},
			"storage_profile": schema.StringAttribute{
				MarkdownDescription: "Storage Profile ID",
				Computed:            true,
			},
		},
	}
}

func (d *WarehouseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*util.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *WarehouseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WarehouseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := data.Id.ValueString()
	warehouse, _, err := d.client.V2.DefaultApi.GetWarehouse(ctx, *d.client.OrganizationId, warehouseId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Warehouse not found", err.Error())
		return
	}

	if name, ok := warehouse.GetNameOk(); ok {
		data.Name = types.StringValue(*name)
	} else {
		data.Name = types.StringNull()
	}

	if orgId, ok := warehouse.GetOrganizationIdOk(); ok {
		data.OrganizationId = types.StringValue(*orgId)
	} else {
		data.OrganizationId = types.StringNull()
	}

	if storageProfile, ok := warehouse.GetStorageProfileOk(); ok {
		data.StorageProfile = types.StringValue(*storageProfile)
	} else {
		data.StorageProfile = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
