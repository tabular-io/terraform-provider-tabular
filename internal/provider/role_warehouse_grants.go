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
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/validators"
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
	RoleName            types.String `tfsdk:"role_name"`
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

func (r *roleWarehouseGrantsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleWarehouseGrantsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := state.WarehouseId.ValueString()
	roleName := state.RoleName.ValueString()

	warehouseGrants, _, err := r.client.V2.DefaultApi.GetRoleWarehouseGrants(ctx, *r.client.OrganizationId, warehouseId, roleName).Execute()
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

	state.Id = types.StringValue(fmt.Sprintf("%s.%s", warehouseId, roleName))

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleWarehouseGrantsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleWarehouseGrantsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warehouseId := plan.WarehouseId.ValueString()
	roleName := plan.RoleName.ValueString()

	var planPrivileges, planPrivilegesWithGrant []string

	for _, privilege := range planPrivileges {
		withGrant := false
		_, _, err := r.client.V2.DefaultApi.CreateWarehouseGrant(ctx, *r.client.OrganizationId, warehouseId).
			CreateWarehouseGrantRequest(tabular.CreateWarehouseGrantRequest{
				SubjectId: &roleName,
				Privilege: &privilege,
				WithGrant: &withGrant,
			}).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError("Unable to create grant", "Unable to create grant"+err.Error())
		}
		resp.Diagnostics.Append(plan.Privileges.ElementsAs(ctx, privilege, false)...)
	}

	for _, privilege := range planPrivilegesWithGrant {
		withGrant := false
		_, _, err := r.client.V2.DefaultApi.CreateWarehouseGrant(ctx, *r.client.OrganizationId, warehouseId).
			CreateWarehouseGrantRequest(tabular.CreateWarehouseGrantRequest{
				SubjectId: &roleName,
				Privilege: &privilege,
				WithGrant: &withGrant,
			}).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError("Unable to create grant", "Unable to create grant"+err.Error())
		}
		resp.Diagnostics.Append(plan.Privileges.ElementsAs(ctx, privilege, false)...)
	}

	plan.Id = types.StringValue(fmt.Sprintf("%s.%s", warehouseId, roleName))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleWarehouseGrantsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Warehouse grant update not supported", "A warehouse grant update shouldn't be possible; please file an issue with the maintainers")
}

func (r *roleWarehouseGrantsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state warehouseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	warehouseId := state.Id.ValueString()
	_, err := r.client.V2.DefaultApi.DeleteWarehouse(ctx, *r.client.OrganizationId, warehouseId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting warehouse", "Unable to delete warehouse "+err.Error())
	}
}
