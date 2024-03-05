package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/tabular-sdk-go/tabular"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
)

var (
	_ resource.Resource                = &serviceAccountResource{}
	_ resource.ResourceWithConfigure   = &serviceAccountResource{}
	_ resource.ResourceWithImportState = &serviceAccountResource{}
)

type serviceAccountResource struct {
	client *util.Client
}

func NewServiceAccountResource() resource.Resource {
	return &serviceAccountResource{}
}

type serviceAccountResourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	RoleId           types.String `tfsdk:"role_id"`
	CredentialKey    types.String `tfsdk:"credential_key"`
	CredentialSecret types.String `tfsdk:"credential_secret"`
}

func (r *serviceAccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*util.Client)
}

func (r *serviceAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *serviceAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Tabular Service Account",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Service account name",
				Required:    true,
			},
			"role_id": schema.StringAttribute{
				Description: "Role ID",
				Required:    true,
			},
			"credential_key": schema.StringAttribute{
				Description: "Credential ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"credential_secret": schema.StringAttribute{
				Description: "Credential secret",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *serviceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := serviceAccountResourceModel{
		Id: types.StringValue(req.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serviceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceAccountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialKey := state.Id.ValueString()
	serviceAccount, _, err := r.client.V2.DefaultAPI.GetCredential(ctx, *r.client.OrganizationId, credentialKey).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read service account", "Unable to read service account "+err.Error())
	}

	state.CredentialKey = types.StringValue(credentialKey)

	if credentialSecret, ok := serviceAccount.GetEncodedSecretOk(); ok {
		state.CredentialSecret = types.StringValue(*credentialSecret)
	}

	if name, ok := serviceAccount.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if roleId, ok := serviceAccount.GetRoleIdOk(); ok {
		state.RoleId = types.StringValue(*roleId)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccountName := plan.Name.ValueString()
	roleId := plan.RoleId.ValueString()

	serviceAccountResponse, _, err := r.client.V2.DefaultAPI.CreateServiceAccountCredential(ctx, *r.client.OrganizationId).
		CreateServiceAccountCredentialRequest(tabular.CreateServiceAccountCredentialRequest{
			Name:   &serviceAccountName,
			RoleId: &roleId,
		}).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Error creating service account", "Unable to create service account "+err.Error())
	}

	if id, ok := serviceAccountResponse.GetCredentialIdOk(); ok {
		plan.Id = types.StringValue(*id)
		plan.CredentialKey = types.StringValue(*id)
	} else {
		resp.Diagnostics.AddError("Unable to set credential_key", "Unable to set credential_key")
	}

	if credentialSecret, ok := serviceAccountResponse.GetCredentialSecretOk(); ok {
		plan.CredentialSecret = types.StringValue(*credentialSecret)
	} else {
		resp.Diagnostics.AddError("Unable to set credential_secret", "Unable to set credential_secret")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *serviceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Service Account Update Not Supported", "A service_account update shouldn't be possible; please file an issue with the maintainers")
}

func (r *serviceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	_, err := r.client.V2.DefaultAPI.DeleteServiceAccountCredential(ctx, *r.client.OrganizationId, state.CredentialKey.ValueString()).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Error deleting serviceAccount", "Unable to delete serviceAccount "+err.Error())
	}
}
