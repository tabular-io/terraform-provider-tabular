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
	"github.com/tabular-io/terraform-provider-tabular/internal"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
)

var (
	_ resource.Resource              = &roleWarehouseGrantsResource{}
	_ resource.ResourceWithConfigure = &roleWarehouseGrantsResource{}
)

type roleWarehouseGrantsResource struct {
	client *util.Client
}

func NewRoleWarehouseGrantsResource() resource.Resource {
	return &roleWarehouseGrantsResource{}
}

type roleWarehouseGrantsResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	RoleId              types.String `tfsdk:"role_id"`
	WarehouseId         types.String `tfsdk:"warehouse_id"`
	Privileges          types.Set    `tfsdk:"privileges"`
	PrivilegesWithGrant types.Set    `tfsdk:"privileges_with_grant"`
}

func (r *roleWarehouseGrantsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*util.Client)
}

func (r *roleWarehouseGrantsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_warehouse_grants"
}

func (r *roleWarehouseGrantsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Managed the grants a role has for a warehouse",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform resource id",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_id": schema.StringAttribute{
				Description: "Role UUID",
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
			"privileges": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Allowed Values: MODIFY_WAREHOUSE, LIST_DATABASES, CREATE_DATABASE, FUTURE_MODIFY_DATABASE, FUTURE_LIST_TABLES, FUTURE_CREATE_TABLE, FUTURE_SELECT, FUTURE_UPDATE, FUTURE_MODIFY_TABLE",
			},
			"privileges_with_grant": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Allowed Values: MODIFY_WAREHOUSE, LIST_DATABASES, CREATE_DATABASE, FUTURE_MODIFY_DATABASE, FUTURE_LIST_TABLES, FUTURE_CREATE_TABLE, FUTURE_SELECT, FUTURE_UPDATE, FUTURE_MODIFY_TABLE",
			},
		},
	}
}

func (r *roleWarehouseGrantsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleWarehouseGrantsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.WarehouseId.ValueString()
	roleId := state.RoleId.ValueString()

	warehouseGrants, _, err := r.client.V2.DefaultApi.GetRoleWarehouseGrants(ctx, *r.client.OrganizationId, warehouseId, roleId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error getting role grants", "Could not get grants for warehouse "+err.Error())
	}

	var privileges []string
	var privilegesWithGrant []string

	for _, grant := range warehouseGrants.Authorizations {
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

	state.Id = types.StringValue(fmt.Sprintf("%s/%s", warehouseId, roleId))

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleWarehouseGrantsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleWarehouseGrantsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := plan.WarehouseId.ValueString()
	roleId := plan.RoleId.ValueString()

	var planPrivileges, planPrivilegesWithGrant []string
	resp.Diagnostics.Append(plan.Privileges.ElementsAs(ctx, &planPrivileges, false)...)
	resp.Diagnostics.Append(plan.PrivilegesWithGrant.ElementsAs(ctx, &planPrivilegesWithGrant, false)...)

	var roleWarehouseGrantRequest []tabular.RoleWarehouseGrantRequest
	roleWarehouseGrantRequest = append(
		warehousePrivilegeRequest(planPrivileges, false, roleId),
		warehousePrivilegeRequest(planPrivilegesWithGrant, true, roleId)...)

	_, err := r.client.V2.DefaultApi.GrantPrivilegesOnWarehouse(ctx, *r.client.OrganizationId, warehouseId).
		RoleWarehouseGrantRequest(roleWarehouseGrantRequest).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Unable to create grant", "Unable to create grant"+err.Error())
		return
	}

	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", warehouseId, roleId))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleWarehouseGrantsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleWarehouseGrantsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := plan.WarehouseId.ValueString()
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
		warehousePrivilegeRequest(privilegesToRemove, false, roleId),
		warehousePrivilegeRequest(privilegesToRemoveWithGrant, true, roleId)...)

	if len(privilegesToRemoveRequest) > 0 {
		_, err := r.client.V2.DefaultApi.RevokePrivilegesOnWarehouse(ctx, *r.client.OrganizationId, warehouseId).
			RoleWarehouseGrantRequest(privilegesToRemoveRequest).
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
		warehousePrivilegeRequest(privilegesToAdd, false, roleId),
		warehousePrivilegeRequest(privilegesToAddWithGrant, true, roleId)...)

	if len(privilegesToAddRequest) > 0 {
		_, err := r.client.V2.DefaultApi.GrantPrivilegesOnWarehouse(ctx, *r.client.OrganizationId, warehouseId).
			RoleWarehouseGrantRequest(privilegesToAddRequest).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError("Unable to create grant", "Unable to create grant"+err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleWarehouseGrantsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleWarehouseGrantsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.WarehouseId.ValueString()
	roleId := state.RoleId.ValueString()

	var statePrivileges, statePrivilegesWithGrant []string
	resp.Diagnostics.Append(state.Privileges.ElementsAs(ctx, &statePrivileges, false)...)
	resp.Diagnostics.Append(state.PrivilegesWithGrant.ElementsAs(ctx, &statePrivilegesWithGrant, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var roleWarehouseGrantRequest []tabular.RoleWarehouseGrantRequest
	roleWarehouseGrantRequest = append(
		warehousePrivilegeRequest(statePrivileges, false, roleId),
		warehousePrivilegeRequest(statePrivilegesWithGrant, true, roleId)...)

	_, err := r.client.V2.DefaultApi.RevokePrivilegesOnWarehouse(ctx, *r.client.OrganizationId, warehouseId).RoleWarehouseGrantRequest(roleWarehouseGrantRequest).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Unable to revoke grants", "Unable to revoke grants"+err.Error())
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func warehousePrivilegeRequest(privileges []string, withGrant bool, roleId string) []tabular.RoleWarehouseGrantRequest {
	var roleWarehouseGrantRequest []tabular.RoleWarehouseGrantRequest
	for _, privilege := range privileges {
		// https://go.dev/blog/loopvar-preview
		privilegeCopy := privilege
		roleWarehouseGrantRequest = append(roleWarehouseGrantRequest, tabular.RoleWarehouseGrantRequest{
			RoleId:    &roleId,
			Privilege: &privilegeCopy,
			WithGrant: &withGrant,
		})
	}
	return roleWarehouseGrantRequest
}
