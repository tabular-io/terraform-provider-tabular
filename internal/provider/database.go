package provider

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/tabular-sdk-go/tabular"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
	"strings"
)

var (
	_ resource.Resource                 = &databaseResource{}
	_ resource.ResourceWithConfigure    = &databaseResource{}
	_ resource.ResourceWithImportState  = &databaseResource{}
	_ resource.ResourceWithUpgradeState = &databaseResource{}
)

type databaseResource struct {
	client *util.Client
}

func NewDatabaseResource() resource.Resource {
	return &databaseResource{}
}

type databaseResourceModelV0 struct {
	WarehouseId types.String `tfsdk:"warehouse_id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
}

type databaseResourceModelV1 struct {
	Id          types.String `tfsdk:"id"`
	WarehouseId types.String `tfsdk:"warehouse_id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
}

type databaseResourceModel struct {
	Id          types.String `tfsdk:"id"`
	WarehouseId types.String `tfsdk:"warehouse_id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
}

func (r *databaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*util.Client)
}

func (r *databaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *databaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "A Tabular Database",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Database ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"warehouse_id": schema.StringAttribute{
				Description: "Warehouse ID (uuid)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Database Name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"location": schema.StringAttribute{
				Description: "Storage Location",
				Computed:    true,
			},
		},
	}
}

func (r *databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Could not parse ", "Expected warehouseId/databaseId")
		return
	}
	warehouseId := parts[0]
	databaseId := parts[1]

	_, warehouseIdErr := uuid.Parse(warehouseId)
	if warehouseIdErr != nil {
		resp.Diagnostics.AddError("Invalid Warehouse ID", warehouseIdErr.Error())
	}
	_, databaseIdErr := uuid.Parse(databaseId)
	if databaseIdErr != nil {
		resp.Diagnostics.AddError("Invalid Database ID", databaseIdErr.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state := databaseResourceModel{
		WarehouseId: types.StringValue(warehouseId),
		Id:          types.StringValue(databaseId),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *databaseResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Add Id attribute to database v0 resources
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"warehouse_id": schema.StringAttribute{
						Description: "Warehouse ID (uuid)",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Database Name",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"location": schema.StringAttribute{
						Description: "Storage Location",
						Computed:    true,
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData databaseResourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)

				if resp.Diagnostics.HasError() {
					return
				}
				retryFunc := util.RetryResourceResponse[*tabular.GetDatabaseResponse]
				databaseResp, _, err := retryFunc(r.client.V2.DefaultAPI.GetDatabase(ctx, *r.client.OrganizationId,
					priorStateData.WarehouseId.ValueString(),
					priorStateData.Name.ValueString()).Execute)

				if err != nil {
					resp.Diagnostics.AddError("Unable to fetch database id for %s", priorStateData.Name.ValueString())
					return
				}

				upgradedStateData := databaseResourceModelV1{
					Id:          types.StringValue(*databaseResp.Id),
					WarehouseId: priorStateData.WarehouseId,
					Name:        priorStateData.Name,
					Location:    priorStateData.Location,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}

func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.WarehouseId.ValueString()
	databaseId := state.Id.ValueString()
	retryFunc := util.RetryResourceResponse[*tabular.GetDatabaseResponse]
	database, httpResponse, err := retryFunc(r.client.V2.DefaultAPI.GetDatabase(ctx, *r.client.OrganizationId, warehouseId, databaseId).Type_("id").Execute)
	if err != nil && !(httpResponse != nil && httpResponse.StatusCode == 404) {
		resp.Diagnostics.AddError(
			"Error fetching database",
			fmt.Sprintf("Could not fetch database %s in warehouse %s: %s", databaseId, warehouseId, err.Error()),
		)
		return
	}
	if database == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(*database.Name)
	value, ok := (*database.Properties)["location"]
	if !ok {
		resp.Diagnostics.AddError("Database in unexpected state", "Database did not have location table property set")
		return
	}
	state.Location = types.StringValue(value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	warehouseId := plan.WarehouseId.ValueString()

	db, _, err := r.client.V2.DefaultAPI.CreateDatabase(ctx, *r.client.OrganizationId, warehouseId).
		CreateDatabaseRequest(tabular.CreateDatabaseRequest{Name: &name}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error creating database", "Could not create database: "+err.Error())
		return
	}

	value, ok := db.GetIdOk()
	if ok {
		plan.Id = types.StringValue(*value)
	} else {
		resp.Diagnostics.AddError("Unable to get database id", "Unable to get database id")
	}

	propertiesMap, ok := db.GetPropertiesOk()
	if ok {
		props := *propertiesMap
		plan.Location = types.StringValue(props["location"])
	} else {
		resp.Diagnostics.AddWarning("Database in unexpected state", "Database did not have location table property set")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Database Update Not Supported", "A database update shouldn't be possible; please file an issue with the maintainers")
}

func (r *databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data databaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	databaseId := data.Id.ValueString()
	warehouseId := data.WarehouseId.ValueString()

	_, err := util.RetryResponse(r.client.V2.DefaultAPI.DeleteDatabase(ctx, *r.client.OrganizationId, warehouseId, databaseId).Execute)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting database", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
