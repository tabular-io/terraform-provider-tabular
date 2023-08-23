package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/validators"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
	"strings"
)

var (
	_ resource.Resource                = &roleDatabaseGrantsResource{}
	_ resource.ResourceWithConfigure   = &roleDatabaseGrantsResource{}
	_ resource.ResourceWithImportState = &roleDatabaseGrantsResource{}
	_ resource.ResourceWithModifyPlan  = &roleDatabaseGrantsResource{}
)

type roleDatabaseGrantsResource struct {
	client *tabular.Client
}

func (r *roleDatabaseGrantsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*tabular.Client)
}

func NewRoleDatabaseGrantsResource() resource.Resource {
	return &roleDatabaseGrantsResource{}
}

type roleDatabaseGrantsModel struct {
	RoleName            types.String `tfsdk:"role_name"`
	WarehouseId         types.String `tfsdk:"warehouse_id"`
	Database            types.String `tfsdk:"database"`
	Privileges          types.Set    `tfsdk:"privileges"`
	PrivilegesWithGrant types.Set    `tfsdk:"privileges_with_grant"`
}

func (r *roleDatabaseGrantsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_database_grants"
}

func (r *roleDatabaseGrantsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the grants a role has for a database.",
		Attributes: map[string]schema.Attribute{
			"role_name": schema.StringAttribute{
				Description: "Role Name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"warehouse_id": schema.StringAttribute{
				Description: "Warehouse ID (uuid)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database": schema.StringAttribute{
				Description: "Database Name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"privileges": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators:  []validator.Set{validators.PrivilegeSetValidator{}},
				Description: "Allowed Values: CREATE_TABLE, LIST_TABLES, MODIFY_DATABASE, FUTURE_SELECT, FUTURE_UPDATE, FUTURE_DROP_TABLE",
			},
			"privileges_with_grant": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators:  []validator.Set{validators.PrivilegeSetValidator{}},
				Description: "Allowed Values: CREATE_TABLE, LIST_TABLES, MODIFY_DATABASE, FUTURE_SELECT, FUTURE_UPDATE, FUTURE_DROP_TABLE",
			},
		},
	}
}

func (r *roleDatabaseGrantsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid role database grant specifier", "Expected three part value, split by a /")
	}
	state := roleDatabaseGrantsModel{
		WarehouseId:         types.StringValue(parts[0]),
		Database:            types.StringValue(parts[1]),
		RoleName:            types.StringValue(parts[2]),
		Privileges:          types.SetUnknown(types.StringType),
		PrivilegesWithGrant: types.SetUnknown(types.StringType),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *roleDatabaseGrantsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleDatabaseGrantsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.WarehouseId.ValueString()
	database := state.Database.ValueString()
	roleName := state.RoleName.ValueString()
	grants, err := r.client.GetRoleDatabaseGrants(warehouseId, database, roleName)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching grants for role", err.Error())
		return
	}
	if grants == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if grants.Privileges == nil {
		state.Privileges = types.SetNull(types.StringType)
	} else {
		state.Privileges, diags = types.SetValueFrom(ctx, types.StringType, grants.Privileges)
		resp.Diagnostics.Append(diags...)
	}

	if grants.PrivilegesWithGrant == nil {
		state.PrivilegesWithGrant = types.SetNull(types.StringType)
	} else {
		state.PrivilegesWithGrant, diags = types.SetValueFrom(ctx, types.StringType, grants.PrivilegesWithGrant)
		resp.Diagnostics.Append(diags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *roleDatabaseGrantsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleDatabaseGrantsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := plan.WarehouseId.ValueString()
	database := plan.Database.ValueString()
	roleName := plan.RoleName.ValueString()

	var planPrivs, planPrivsWithGrant []string
	resp.Diagnostics.Append(plan.Privileges.ElementsAs(ctx, &planPrivs, false)...)
	resp.Diagnostics.Append(plan.PrivilegesWithGrant.ElementsAs(ctx, &planPrivsWithGrant, false)...)

	err := r.client.AddRoleDatabaseGrants(warehouseId, database, roleName, planPrivsWithGrant, true)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("privileges_with_grant"),
			"Failure adding privileges with grant",
			err.Error(),
		)
	}
	err = r.client.AddRoleDatabaseGrants(warehouseId, database, roleName, planPrivs, false)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("privileges"),
			"Failure adding privileges",
			err.Error(),
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleDatabaseGrantsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleDatabaseGrantsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := plan.WarehouseId.ValueString()
	database := plan.Database.ValueString()
	roleName := plan.RoleName.ValueString()

	var planPrivs, statePrivs, planPrivsWithGrant, statePrivsWithGrant []string
	resp.Diagnostics.Append(plan.Privileges.ElementsAs(ctx, &planPrivs, false)...)
	resp.Diagnostics.Append(plan.PrivilegesWithGrant.ElementsAs(ctx, &planPrivsWithGrant, false)...)
	resp.Diagnostics.Append(state.Privileges.ElementsAs(ctx, &statePrivs, false)...)
	resp.Diagnostics.Append(state.PrivilegesWithGrant.ElementsAs(ctx, &statePrivsWithGrant, false)...)

	withGrantToRemove := internal.Difference(statePrivsWithGrant, planPrivsWithGrant)
	toRemove := internal.Difference(statePrivs, planPrivs)
	withGrantToAdd := internal.Difference(planPrivsWithGrant, statePrivsWithGrant)
	toAdd := internal.Difference(planPrivs, statePrivs)

	err := r.client.RevokeRoleDatabaseGrants(warehouseId, database, roleName, withGrantToRemove, true)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("privileges_with_grant"),
			"Failure removing privileges with grant",
			err.Error(),
		)
	}
	err = r.client.RevokeRoleDatabaseGrants(warehouseId, database, roleName, toRemove, false)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("privileges"),
			"Failure removing privileges",
			err.Error(),
		)
	}
	err = r.client.AddRoleDatabaseGrants(warehouseId, database, roleName, withGrantToAdd, true)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("privileges_with_grant"),
			"Failure adding privileges with grant",
			err.Error(),
		)
	}
	err = r.client.AddRoleDatabaseGrants(warehouseId, database, roleName, toAdd, false)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("privileges"),
			"Failure adding privileges",
			err.Error(),
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleDatabaseGrantsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleDatabaseGrantsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.WarehouseId.ValueString()
	database := state.Database.ValueString()
	roleName := state.RoleName.ValueString()

	var statePrivs, statePrivsWithGrant []string
	resp.Diagnostics.Append(state.Privileges.ElementsAs(ctx, &statePrivs, false)...)
	resp.Diagnostics.Append(state.PrivilegesWithGrant.ElementsAs(ctx, &statePrivsWithGrant, false)...)

	err := r.client.RevokeRoleDatabaseGrants(warehouseId, database, roleName, statePrivsWithGrant, true)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("privileges_with_grant"),
			"Failure removing privileges with grant",
			err.Error(),
		)
	}
	err = r.client.RevokeRoleDatabaseGrants(warehouseId, database, roleName, statePrivs, false)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("privileges"),
			"Failure removing privileges",
			err.Error(),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *roleDatabaseGrantsResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan roleDatabaseGrantsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Treat an empty list, null, and (known after apply) for privileges (w/ or w/o grant) interchangeably
	if plan.Privileges.IsUnknown() || plan.Privileges.IsNull() {
		plan.Privileges = types.SetValueMust(types.StringType, []attr.Value{})
	}
	if plan.PrivilegesWithGrant.IsUnknown() || plan.PrivilegesWithGrant.IsNull() {
		plan.PrivilegesWithGrant = types.SetValueMust(types.StringType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}
