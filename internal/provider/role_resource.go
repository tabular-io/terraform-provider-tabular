package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ExternalId types.String `tfsdk:"external_id"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"external_id": schema.StringAttribute{
				Description: "External ID",
				Computed:    true,
				Optional:    true,
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

	roleId := state.Id.ValueString()
	role, err := r.client.GetRole(roleId)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching role", "Could not get fetch role with ID "+roleId+": "+err.Error())
		return
	}
	if role == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(role.Name)
	state.ExternalId = types.StringNull()
	diags = resp.State.Set(ctx, &state)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.CreateRole(plan.Name.ValueString())
	tflog.Warn(ctx, "Role Name: "+role.Name)
	if err != nil {
		resp.Diagnostics.AddError("Error creating role", "Could not create role: "+err.Error())
	}

	plan.Id = types.StringValue(role.Id)
	if role.ExternalId != nil {
		plan.ExternalId = types.StringValue(*role.ExternalId)
	} else {
		plan.ExternalId = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Role updates not supported", "...how did you get here?")
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	roleId := data.Id.ValueString()

	err := r.client.DeleteRole(roleId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting role", "Something went wrong. Does the role still have any users/roles/permissions attached to it? Err: "+err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
