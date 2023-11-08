package provider

import (
	"context"
	"fmt"
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
	_ resource.Resource                = &databaseResource{}
	_ resource.ResourceWithConfigure   = &databaseResource{}
	_ resource.ResourceWithImportState = &databaseResource{}
)

type databaseResource struct {
	client *util.Client
}

func NewDatabaseResource() resource.Resource {
	return &databaseResource{}
}

type databaseResourceModel struct {
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
		Description: "A Tabular Database",
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
	}
}

func (r *databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Could not parse ", "Expected two part value, split by a /")
	}
	state := databaseResourceModel{
		WarehouseId: types.StringValue(parts[0]),
		Name:        types.StringValue(parts[1]),
		Location:    types.StringUnknown(),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.WarehouseId.ValueString()
	databaseName := state.Name.ValueString()
	database, _, err := r.client.V2.DefaultApi.GetDatabase(ctx, *r.client.OrganizationId, warehouseId, databaseName).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching database",
			fmt.Sprintf("Could not fetch database %s in warehouse %s: %s", databaseName, warehouseId, err.Error()),
		)
		return
	}
	if database.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

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

	_, _, err := r.client.V2.DefaultApi.CreateDatabase(ctx, *r.client.OrganizationId, warehouseId).
		CreateDatabaseRequest(tabular.CreateDatabaseRequest{Name: &name}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error creating database", "Could not create database: "+err.Error())
		return
	}

	db, _, err := r.client.V2.DefaultApi.GetDatabase(ctx, *r.client.OrganizationId, warehouseId, name).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching database",
			fmt.Sprintf("Could not fetch database %s in warehouse %s: %s", name, *db.WarehouseId, err.Error()),
		)
		return
	}

	value, ok := (*db.Properties)["location"]
	if ok {
		plan.Location = types.StringValue(value)
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

	database := data.Name.ValueString()
	warehouseId := data.WarehouseId.ValueString()

	err := r.client.V1.DeleteDatabase(warehouseId, database)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting database", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
