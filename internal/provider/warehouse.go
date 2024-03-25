package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/tabular-sdk-go/tabular"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
)

var (
	_ resource.Resource                = &warehouseResource{}
	_ resource.ResourceWithConfigure   = &warehouseResource{}
	_ resource.ResourceWithImportState = &warehouseResource{}
)

type warehouseResource struct {
	client *util.Client
}

func NewWarehouseResource() resource.Resource {
	return &warehouseResource{}
}

type warehouseResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	StorageProfile types.String `tfsdk:"storage_profile"`
}

func (r *warehouseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*util.Client)
}

func (r *warehouseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_warehouse"
}

func (r *warehouseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Tabular Warehouse",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Warehouse ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Warehouse name",
				Required:    true,
			},
			"storage_profile": schema.StringAttribute{
				Description: "Storage profile",
				Required:    true,
			},
		},
	}
}

func (r *warehouseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := warehouseResourceModel{
		Id: types.StringValue(req.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *warehouseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state warehouseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.Id.ValueString()
	warehouse, _, err := util.RetryResourceResponse[*tabular.GetWarehouseResponse](
		r.client.V2.DefaultAPI.GetWarehouse(ctx, *r.client.OrganizationId, warehouseId).Execute,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error getting warehouse", "Could not get warehouse "+err.Error())
		return
	}

	if storageProfileId, ok := warehouse.GetStorageProfileOk(); ok {
		state.StorageProfile = types.StringValue(*storageProfileId)
	} else {
		state.StorageProfile = types.StringNull()
	}

	if warehouseName, ok := warehouse.GetNameOk(); ok {
		state.Name = types.StringValue(*warehouseName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *warehouseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan warehouseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseName := plan.Name.ValueString()
	storageProfileId := plan.StorageProfile.ValueString()

	apiCreateWarehouseRequest := r.client.V2.DefaultAPI.CreateWarehouse(ctx, *r.client.OrganizationId)
	warehouseResponse, _, err := util.RetryResourceResponse[*tabular.CreateWarehouseResponse](apiCreateWarehouseRequest.
		CreateWarehouseRequest(tabular.CreateWarehouseRequest{
			Name:             &warehouseName,
			StorageProfileId: &storageProfileId,
		}).Execute)

	if err != nil {
		resp.Diagnostics.AddError("Error creating storage profile", "Unable to create storage profile "+err.Error())
		return
	}

	if id, ok := warehouseResponse.GetIdOk(); ok {
		plan.Id = types.StringValue(*id)
	} else {
		resp.Diagnostics.AddError("Unable to set storage profile id", "Unable to set storage profile id")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *warehouseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Database Update Not Supported", "A database update shouldn't be possible; please file an issue with the maintainers")
}

func (r *warehouseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state warehouseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	warehouseId := state.Id.ValueString()
	_, err := util.RetryResponse(r.client.V2.DefaultAPI.DeleteWarehouse(ctx, *r.client.OrganizationId, warehouseId).Execute)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting warehouse", "Unable to delete warehouse "+err.Error())
	}
}
