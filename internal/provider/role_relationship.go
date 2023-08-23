package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
	"strings"
)

var (
	_ resource.Resource                = &roleRelationshipResource{}
	_ resource.ResourceWithConfigure   = &roleRelationshipResource{}
	_ resource.ResourceWithImportState = &roleRelationshipResource{}
)

type roleRelationshipResource struct {
	client *util.Client
}

func NewRoleRelationshipResource() resource.Resource {
	return &roleRelationshipResource{}
}

func (r *roleRelationshipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*util.Client)
}

type roleRelationshipModel struct {
	ParentRoleName types.String `tfsdk:"parent_role_name"`
	ChildRoleName  types.String `tfsdk:"child_role_name"`
}

func (r *roleRelationshipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_relationship"
}

func (r *roleRelationshipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Relationship between two roles",
		Attributes: map[string]schema.Attribute{
			"parent_role_name": schema.StringAttribute{
				Description: "Parent role name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"child_role_name": schema.StringAttribute{
				Description: "Child role name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *roleRelationshipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid role relationship specifier", "Expected two part value, split by a /")
	}
	state := roleRelationshipModel{
		ParentRoleName: types.StringValue(parts[0]),
		ChildRoleName:  types.StringValue(parts[1]),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *roleRelationshipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleRelationshipModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentRoleName := state.ParentRoleName.ValueString()
	role, err := r.client.V1.GetRole(parentRoleName)
	if role == nil {
		resp.Diagnostics.AddError("Error fetching role", "Could not fetch role "+parentRoleName)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error fetching role", "Could not fetch role "+parentRoleName+": "+err.Error())
		return
	}
	found := false
	childRoleName := state.ChildRoleName.ValueString()
	for _, child := range role.Children {
		if child.Name == childRoleName {
			found = true
		}
	}
	if !found {
		resp.Diagnostics.AddError("Role relationship not found", "Could not find role relationship between "+parentRoleName+" and "+childRoleName)
	}
}

func (r *roleRelationshipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleRelationshipModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.V1.AddRoleRelation(plan.ParentRoleName.ValueString(), plan.ChildRoleName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating role relation", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleRelationshipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Role relation unsupported", "It shouldn't be possible to cause an update")
}

func (r *roleRelationshipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan roleRelationshipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.V1.DeleteRoleRelation(plan.ParentRoleName.ValueString(), plan.ChildRoleName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating role relation", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
