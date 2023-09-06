package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
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
	StorageProfile types.String `tfsdk:"storage_profile"`
	Region         types.String `tfsdk:"region"`
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
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Warehouse Name",
				Optional:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Computed:            true,
			},
			"storage_profile": schema.StringAttribute{
				MarkdownDescription: "Storage Profile ID",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Warehouse Region",
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

	if data.Id.IsNull() {
		getWarehouseByName(*d.client.V1, data, resp.Diagnostics)
	} else {
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

		if region, ok := warehouse.GetRegionOk(); ok {
			data.Region = types.StringValue(*region)
		} else {
			data.Region = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getWarehouseByName(client tabular.Client, data WarehouseDataSourceModel, diag diag.Diagnostics) {
	warehouses, err := client.GetWarehouses()
	if err != nil {
		diag.AddError("Failed fetching warehouses", err.Error())
	}

	targetName := data.Name.ValueString()
	if targetName != "" {
		for _, w := range warehouses {
			if w.Name == targetName {
				data.Id = types.StringValue(w.Id)
				data.Region = types.StringValue(w.Region)
				break
			}
		}
	}

	diag.AddError("Warehouse not found", fmt.Sprintf("Could not find warehouse with name %s", targetName))
}
