package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
)

var _ datasource.DataSource = &WarehouseDataSource{}
var _ datasource.DataSourceWithConfigure = &WarehouseDataSource{}

func NewWarehouseDataSource() datasource.DataSource {
	return &WarehouseDataSource{}
}

type WarehouseDataSource struct {
	client *tabular.Client
}

type WarehouseDataSourceModel struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`
}

func (w *WarehouseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_warehouse"
}

func (w *WarehouseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Warehouse data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Role Name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ID",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Region",
				Computed:            true,
			},
		},
	}
}

func (w *WarehouseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	w.client = client
}

func (w *WarehouseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WarehouseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	warehouses, err := w.client.GetWarehouses()
	if err != nil {
		resp.Diagnostics.AddError("Failed fetching warehouses", err.Error())
		return
	}
	found := false
	targetName := data.Name.ValueString()
	if targetName != "" {
		for _, w := range warehouses {
			if w.Name == targetName {
				data.Id = types.StringValue(w.Id)
				data.Region = types.StringValue(w.Region)
				found = true
				break
			}
		}
	}

	if !found {
		resp.Diagnostics.AddError("Warehouse not found", fmt.Sprintf("Could not find warehouse with name %s", targetName))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
