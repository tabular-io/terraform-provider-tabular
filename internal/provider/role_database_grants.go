package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/tabular-sdk-go/tabular"
	"github.com/tabular-io/terraform-provider-tabular/internal"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/validators"
	"strings"
)

var (
	_ resource.Resource                 = &roleDatabaseGrantsResource{}
	_ resource.ResourceWithConfigure    = &roleDatabaseGrantsResource{}
	_ resource.ResourceWithImportState  = &roleDatabaseGrantsResource{}
	_ resource.ResourceWithUpgradeState = &roleDatabaseGrantsResource{}
)

type roleDatabaseGrantsResource struct {
	client *util.Client
}

func (r *roleDatabaseGrantsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*util.Client)
}

func NewRoleDatabaseGrantsResource() resource.Resource {
	return &roleDatabaseGrantsResource{}
}

type roleDatabaseGrantsModelV0 struct {
	RoleName            types.String `tfsdk:"role_name"`
	WarehouseId         types.String `tfsdk:"warehouse_id"`
	Database            types.String `tfsdk:"database"`
	Privileges          types.Set    `tfsdk:"privileges"`
	PrivilegesWithGrant types.Set    `tfsdk:"privileges_with_grant"`
}

type roleDatabaseGrantsModelV1 struct {
	Id                  types.String `tfsdk:"id"`
	RoleId              types.String `tfsdk:"role_id"`
	WarehouseId         types.String `tfsdk:"warehouse_id"`
	DatabaseId          types.String `tfsdk:"database_id"`
	Privileges          types.Set    `tfsdk:"privileges"`
	PrivilegesWithGrant types.Set    `tfsdk:"privileges_with_grant"`
}

type roleDatabaseGrantsModel struct {
	Id                  types.String `tfsdk:"id"`
	RoleId              types.String `tfsdk:"role_id"`
	WarehouseId         types.String `tfsdk:"warehouse_id"`
	DatabaseId          types.String `tfsdk:"database_id"`
	Privileges          types.Set    `tfsdk:"privileges"`
	PrivilegesWithGrant types.Set    `tfsdk:"privileges_with_grant"`
}

func (r *roleDatabaseGrantsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_database_grants"
}

func (r *roleDatabaseGrantsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages the grants a role has for a database.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform resource id",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_id": schema.StringAttribute{
				Description: "Role Id",
				Required:    true,
			},
			"warehouse_id": schema.StringAttribute{
				Description: "Warehouse ID (uuid)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_id": schema.StringAttribute{
				Description: "Database Id",
				Required:    true,
			},
			"privileges": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.Set{validators.PrivilegeSetValidator{}},
				Description: "Allowed Values: CREATE_TABLE, LIST_TABLES, MODIFY_DATABASE, FUTURE_SELECT, FUTURE_UPDATE, FUTURE_DROP_TABLE, FUTURE_MANAGE_GRANTS_DATABASE, FUTURE_MANAGE_GRANTS_TABLE",
			},
			"privileges_with_grant": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.Set{validators.PrivilegeSetValidator{}},
				Description: "Allowed Values: CREATE_TABLE, LIST_TABLES, MODIFY_DATABASE, FUTURE_SELECT, FUTURE_UPDATE, FUTURE_DROP_TABLE, FUTURE_MANAGE_GRANTS_DATABASE, FUTURE_MANAGE_GRANTS_TABLE",
			},
		},
	}
}

func (r *roleDatabaseGrantsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid role database grant specifier", "Expected warehouseId/databaseId/roleName")
	}
	roleResp, _, err := r.client.V2.DefaultAPI.GetRole(ctx, *r.client.OrganizationId, parts[2]).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch role id for %s", parts[2])
	}

	warehouseId := parts[0]
	databaseId := parts[1]
	roleId := *roleResp.Id

	state := roleDatabaseGrantsModel{
		Id:                  types.StringValue(fmt.Sprintf("%s/%s/%s", warehouseId, databaseId, roleId)),
		WarehouseId:         types.StringValue(warehouseId),
		DatabaseId:          types.StringValue(databaseId),
		RoleId:              types.StringValue(roleId),
		Privileges:          types.SetUnknown(types.StringType),
		PrivilegesWithGrant: types.SetUnknown(types.StringType),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *roleDatabaseGrantsResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Add Id and RoleId attributes to role_database_grants v0 resource
		0: {
			PriorSchema: &schema.Schema{
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
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData roleDatabaseGrantsModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)

				if resp.Diagnostics.HasError() {
					return
				}

				roleResp, _, err := r.client.V2.DefaultAPI.GetRole(ctx, *r.client.OrganizationId, priorStateData.RoleName.ValueString()).Execute()

				if err != nil {
					resp.Diagnostics.AddError("Unable to fetch role id for %s", priorStateData.RoleName.ValueString())
					return
				}

				roleId := roleResp.Id

				databaseResp, _, err := r.client.V2.DefaultAPI.GetDatabase(ctx, *r.client.OrganizationId,
					priorStateData.WarehouseId.ValueString(),
					priorStateData.Database.ValueString()).Execute()

				if err != nil {
					resp.Diagnostics.AddError("Unable to fetch database id for %s", priorStateData.Database.ValueString())
					return
				}

				databaseId := *databaseResp.Id

				upgradedStateData := roleDatabaseGrantsModelV1{
					Id:                  types.StringValue(fmt.Sprintf("%s/%s/%s", priorStateData.WarehouseId, databaseId, *roleId)),
					DatabaseId:          types.StringValue(databaseId),
					RoleId:              types.StringValue(*roleId),
					WarehouseId:         priorStateData.WarehouseId,
					Privileges:          priorStateData.Privileges,
					PrivilegesWithGrant: priorStateData.PrivilegesWithGrant,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}

func (r *roleDatabaseGrantsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleDatabaseGrantsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.WarehouseId.ValueString()
	databaseId := state.DatabaseId.ValueString()
	roleId := state.RoleId.ValueString()
	databaseGrants, _, err := r.client.V2.DefaultAPI.ListDatabaseRoleGrantsForRole(ctx, *r.client.OrganizationId, warehouseId, databaseId, roleId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error fetching grants for role", err.Error())
		return
	}

	var privileges []string
	var privilegesWithGrant []string

	for _, grant := range databaseGrants.Authorizations {
		if *grant.WithGrant {
			privilegesWithGrant = append(privilegesWithGrant, *grant.Privilege)
		} else {
			privileges = append(privileges, *grant.Privilege)
		}
	}

	state.Privileges, diags = types.SetValueFrom(ctx, types.StringType, privileges)
	resp.Diagnostics.Append(diags...)
	state.PrivilegesWithGrant, diags = types.SetValueFrom(ctx, types.StringType, privilegesWithGrant)
	resp.Diagnostics.Append(diags...)

	state.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", warehouseId, databaseId, roleId))

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
	databaseId := plan.DatabaseId.ValueString()
	roleId := plan.RoleId.ValueString()

	var planPrivileges, planPrivilegesWithGrant []string
	resp.Diagnostics.Append(plan.Privileges.ElementsAs(ctx, &planPrivileges, false)...)
	resp.Diagnostics.Append(plan.PrivilegesWithGrant.ElementsAs(ctx, &planPrivilegesWithGrant, false)...)

	var roleDatabaseGrantRequest []tabular.RoleDatabaseGrantRequest
	roleDatabaseGrantRequest = append(
		databasePrivilegeRequest(planPrivileges, false, roleId),
		databasePrivilegeRequest(planPrivilegesWithGrant, true, roleId)...)

	httpResp, err := r.client.V2.DefaultAPI.GrantPrivilegesOnDatabase(ctx, *r.client.OrganizationId, warehouseId, databaseId).
		RoleDatabaseGrantRequest(roleDatabaseGrantRequest).
		Execute()
	if err != nil {
		errResp, _ := tabular.ParseErrorResponse(httpResp.Body)
		resp.Diagnostics.AddError("Error creating database role grant", fmt.Sprintf("Received %s %s", httpResp.Status, errResp.Error.Type))
		return
	}

	plan.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", warehouseId, databaseId, roleId))

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
	databaseId := plan.DatabaseId.ValueString()
	roleId := plan.RoleId.ValueString()

	var planPrivileges, planPrivilegesWithGrant []string
	resp.Diagnostics.Append(plan.Privileges.ElementsAs(ctx, &planPrivileges, false)...)
	resp.Diagnostics.Append(plan.PrivilegesWithGrant.ElementsAs(ctx, &planPrivilegesWithGrant, false)...)

	var statePlanPrivileges, statePlanPrivilegesWithGrant []string
	resp.Diagnostics.Append(state.Privileges.ElementsAs(ctx, &statePlanPrivileges, false)...)
	resp.Diagnostics.Append(state.PrivilegesWithGrant.ElementsAs(ctx, &statePlanPrivilegesWithGrant, false)...)

	// Remove privileges
	privilegesToRemove := internal.Difference(statePlanPrivileges, planPrivileges)
	privilegesToRemoveWithGrant := internal.Difference(statePlanPrivilegesWithGrant, planPrivilegesWithGrant)
	privilegesToRemoveRequest := append(
		databasePrivilegeRequest(privilegesToRemove, false, roleId),
		databasePrivilegeRequest(privilegesToRemoveWithGrant, true, roleId)...)

	if len(privilegesToRemoveRequest) > 0 {
		_, err := r.client.V2.DefaultAPI.RevokePrivilegesOnDatabase(ctx, *r.client.OrganizationId, warehouseId, databaseId).
			RoleDatabaseGrantRequest(privilegesToRemoveRequest).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError("Unable to revoke grant", "Unable to revoke grant"+err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	// Add privileges
	privilegesToAdd := internal.Difference(planPrivileges, statePlanPrivileges)
	privilegesToAddWithGrant := internal.Difference(planPrivilegesWithGrant, statePlanPrivilegesWithGrant)
	privilegesToAddRequest := append(
		databasePrivilegeRequest(privilegesToAdd, false, roleId),
		databasePrivilegeRequest(privilegesToAddWithGrant, true, roleId)...)

	if len(privilegesToAddRequest) > 0 {
		_, err := r.client.V2.DefaultAPI.GrantPrivilegesOnDatabase(ctx, *r.client.OrganizationId, warehouseId, databaseId).
			RoleDatabaseGrantRequest(privilegesToAddRequest).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError("Unable to create grant", "Unable to create grant"+err.Error())
			return
		}
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
	databaseId := state.DatabaseId.ValueString()
	roleId := state.RoleId.ValueString()

	var statePrivileges, statePrivilegesWithGrant []string
	resp.Diagnostics.Append(state.Privileges.ElementsAs(ctx, &statePrivileges, false)...)
	resp.Diagnostics.Append(state.PrivilegesWithGrant.ElementsAs(ctx, &statePrivilegesWithGrant, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var roleWarehouseGrantRequest []tabular.RoleDatabaseGrantRequest
	roleWarehouseGrantRequest = append(
		databasePrivilegeRequest(statePrivileges, false, roleId),
		databasePrivilegeRequest(statePrivilegesWithGrant, true, roleId)...)

	_, err := r.client.V2.DefaultAPI.RevokePrivilegesOnDatabase(ctx, *r.client.OrganizationId, warehouseId, databaseId).
		RoleDatabaseGrantRequest(roleWarehouseGrantRequest).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Unable to revoke grants", "Unable to revoke grants"+err.Error())
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func databasePrivilegeRequest(privileges []string, withGrant bool, roleId string) []tabular.RoleDatabaseGrantRequest {
	var roleDatabaseGrantRequest []tabular.RoleDatabaseGrantRequest
	for _, privilege := range privileges {
		// https://go.dev/blog/loopvar-preview
		privilegeCopy := privilege
		roleDatabaseGrantRequest = append(roleDatabaseGrantRequest, tabular.RoleDatabaseGrantRequest{
			RoleId:    &roleId,
			Privilege: &privilegeCopy,
			WithGrant: &withGrant,
		})
	}
	return roleDatabaseGrantRequest
}
