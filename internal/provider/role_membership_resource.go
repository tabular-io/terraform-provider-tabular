package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
)

var (
	_ resource.Resource                = &roleMembershipResource{}
	_ resource.ResourceWithConfigure   = &roleMembershipResource{}
	_ resource.ResourceWithImportState = &roleMembershipResource{}
)

type roleMembershipResource struct {
	client *tabular.Client
}

func NewroleMembershipResource() resource.Resource {
	return &roleMembershipResource{}
}

func (r *roleMembershipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*tabular.Client)
}

type roleMembershipModel struct {
	RoleName     types.String `tfsdk:"role_name"`
	AdminMembers types.Set    `tfsdk:"admin_members"`
	Members      types.Set    `tfsdk:"members"`
}

func (r *roleMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_membership"
}

func (r *roleMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Grant users access to a role",
		Attributes: map[string]schema.Attribute{
			"role_name": schema.StringAttribute{
				Description: "Role name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_members": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"members": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *roleMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := roleMembershipModel{
		RoleName:     types.StringValue(req.ID),
		AdminMembers: types.SetUnknown(types.StringType),
		Members:      types.SetUnknown(types.StringType),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *roleMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleMembershipModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleName := state.RoleName.ValueString()
	role, err := r.client.GetRole(roleName)
	if role == nil {
		resp.Diagnostics.AddError("Error fetching role", "Could not fetch role "+roleName)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error fetching role", "Could not fetch role "+roleName+": "+err.Error())
		return
	}

	members := internal.Map(role.Members, func(m tabular.Member) string { return m.Email })
	// TODO: Need update to endpoints exposing withGrant
	state.AdminMembers, diags = types.SetValue(types.StringType, []attr.Value{})
	resp.Diagnostics.Append(diags...)
	state.Members, diags = types.SetValueFrom(ctx, types.StringType, members)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleMembershipModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var adminMemberEmails, memberEmails []string
	resp.Diagnostics.Append(plan.AdminMembers.ElementsAs(ctx, &adminMemberEmails, false)...)
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &memberEmails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgMemberMap, err := r.client.GetOrgMemberIdsMap()
	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch org members", err.Error())
		return
	}
	adminMemberIds := mapMemberEmailsToIds(adminMemberEmails, orgMemberMap, func(email string) {
		resp.Diagnostics.AddAttributeError(
			path.Root("admin_members"),
			"Error adding user",
			fmt.Sprintf("Could not find user with email %s in org", email),
		)
	})
	memberIds := mapMemberEmailsToIds(memberEmails, orgMemberMap, func(email string) {
		resp.Diagnostics.AddAttributeError(
			path.Root("members"),
			"Error adding user",
			fmt.Sprintf("Could not find user with email %s in org", email),
		)
	})

	err = r.client.AddRoleMembers(plan.RoleName.ValueString(), adminMemberIds, memberIds)
	if err != nil {
		resp.Diagnostics.AddError("Error adding role members", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleMembershipModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var planAdminMemberEmails, planMemberEmails, stateAdminMemberEmails, stateMemberEmails []string
	resp.Diagnostics.Append(plan.AdminMembers.ElementsAs(ctx, &planAdminMemberEmails, false)...)
	resp.Diagnostics.Append(state.AdminMembers.ElementsAs(ctx, &stateAdminMemberEmails, false)...)
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &planMemberEmails, false)...)
	resp.Diagnostics.Append(state.Members.ElementsAs(ctx, &stateMemberEmails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgMemberMap, err := r.client.GetOrgMemberIdsMap()
	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch org members", err.Error())
		return
	}
	planAdminMemberIds := mapMemberEmailsToIds(planAdminMemberEmails, orgMemberMap, func(email string) {
		resp.Diagnostics.AddAttributeError(
			path.Root("admin_members"),
			"Error adding user",
			fmt.Sprintf("Could not find user with email %s in org", email),
		)
	})
	stateAdminMemberIds := mapMemberEmailsToIds(stateAdminMemberEmails, orgMemberMap, func(email string) {
		resp.Diagnostics.AddAttributeError(
			path.Root("admin_members"),
			"Error adding user",
			fmt.Sprintf("Could not find user with email %s in org", email),
		)
	})
	planMemberIds := mapMemberEmailsToIds(planMemberEmails, orgMemberMap, func(email string) {
		resp.Diagnostics.AddAttributeError(
			path.Root("members"),
			"Error adding user",
			fmt.Sprintf("Could not find user wit`h email %s in org", email),
		)
	})
	stateMemberIds := mapMemberEmailsToIds(stateMemberEmails, orgMemberMap, func(email string) {
		resp.Diagnostics.AddAttributeError(
			path.Root("members"),
			"Error adding user",
			fmt.Sprintf("Could not find user with email %s in org", email),
		)
	})
	if resp.Diagnostics.HasError() {
		return
	}

	adminToRemove := internal.Difference(stateAdminMemberIds, planAdminMemberIds)
	toRemove := internal.Difference(stateMemberIds, planMemberIds)
	err = r.client.DeleteRoleMembers(state.RoleName.ValueString(), append(adminToRemove, toRemove...))
	if err != nil {
		resp.Diagnostics.AddError("Error removing role members", err.Error())
		return
	}

	adminToAdd := internal.Difference(planAdminMemberIds, stateAdminMemberIds)
	toAdd := internal.Difference(planMemberIds, stateMemberIds)
	err = r.client.AddRoleMembers(state.RoleName.ValueString(), adminToAdd, toAdd)
	if err != nil {
		resp.Diagnostics.AddError("Error adding role members", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var memberEmails []string
	resp.Diagnostics.Append(state.Members.ElementsAs(ctx, &memberEmails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgMemberMap, err := r.client.GetOrgMemberIdsMap()
	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch org members", err.Error())
		return
	}
	memberIds := mapMemberEmailsToIds(memberEmails, orgMemberMap, func(email string) {
		resp.Diagnostics.AddAttributeError(
			path.Root("members"),
			"Error removing user",
			fmt.Sprintf("Could not find user with email %s in org", email),
		)
	})

	err = r.client.DeleteRoleMembers(state.RoleName.ValueString(), memberIds)
	if err != nil {
		resp.Diagnostics.AddError("Error creating role relation", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func mapMemberEmailsToIds(memberEmails []string, memberIdMap map[string]string, errorHandler func(string)) []string {
	memberIds := make([]string, 0, len(memberEmails))
	for _, email := range memberEmails {
		if val, ok := memberIdMap[email]; ok {
			memberIds = append(memberIds, val)
		} else {
			errorHandler(email)
		}
	}
	return memberIds
}
