package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
)

var (
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

type roleResource struct {
	client *tabular.Client
}

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ForceDestroy types.Bool   `tfsdk:"force_destroy"`
}

func (r *roleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Tabular Role",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Role ID (uuid)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Role Name",
				Required:    true,
			},
			"force_destroy": schema.BoolAttribute{
				Description: "Boolean that indicates the role should be destroyed even if it still has associations (e.g." +
					"user assignments, relations to other roles, etc). Defaults to false.",
				Optional: true,
			},
		},
	}
}

func (r *roleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*tabular.Client)
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleName := state.Name.ValueString()
	role, err := r.client.GetRole(roleName)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching role", "Could not get fetch role "+roleName+": "+err.Error())
		return
	}
	if role == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(role.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.CreateRole(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating role", "Could not create role: "+err.Error())
		return
	}

	plan.Id = types.StringValue(role.Id)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var current, target roleResourceModel
	diags := req.State.Get(ctx, &current)
	resp.Diagnostics.Append(diags...)
	diags = req.Plan.Get(ctx, &target)
	resp.Diagnostics.Append(diags...)

	currentName := current.Name.ValueString()
	targetName := target.Name.ValueString()
	if currentName != targetName {
		role, err := r.client.RenameRole(currentName, targetName)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				fmt.Sprintf("Was unable to rename role %s to %s", currentName, targetName),
				err.Error(),
			)
			return
		}
		current.Id = types.StringValue(role.Id)
		current.Name = types.StringValue(role.Name)
	}

	current.ForceDestroy = target.ForceDestroy

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.DeleteRole(data.Name.ValueString(), data.ForceDestroy.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting role", "Something went wrong. Does the role still have any users/roles/permissions attached to it? Err: "+err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
