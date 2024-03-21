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
)

var (
	_ resource.Resource              = &awsRoleMappingResource{}
	_ resource.ResourceWithConfigure = &awsRoleMappingResource{}
)

type awsRoleMappingResource struct {
	client *util.Client
}

func NewAWSRoleMappingResource() resource.Resource {
	return &awsRoleMappingResource{}
}

type awsRoleMappingResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RoleId     types.String `tfsdk:"role_id"`
	AWSRoleArn types.String `tfsdk:"aws_role_arn"`
}

func (r *awsRoleMappingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*util.Client)
}

func (r *awsRoleMappingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_role_mapping"
}

func (r *awsRoleMappingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Tabular AWS Role Mapping",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "",
				Computed:    true,
			},
			"role_id": schema.StringAttribute{
				Description: "Role ID",
				Required:    true,
			},
			"aws_role_arn": schema.StringAttribute{
				Description: "AWS IAM Role ARN",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *awsRoleMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state awsRoleMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialKey := state.Id.ValueString()
	roleMappingAWS, _, err := r.client.V2.DefaultAPI.GetCredential(ctx, *r.client.OrganizationId, credentialKey).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Unable to read service account", "Unable to read service account "+err.Error())
		return
	}

	state.Id = types.StringValue(credentialKey)

	if name, ok := roleMappingAWS.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if roleId, ok := roleMappingAWS.GetRoleIdOk(); ok {
		state.RoleId = types.StringValue(*roleId)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *awsRoleMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan awsRoleMappingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleId := plan.RoleId.ValueString()
	awsRoleArn := plan.AWSRoleArn.ValueString()
	name := fmt.Sprintf("%s-%s", roleId, awsRoleArn)

	roleMappingAWSResponse, _, err := r.client.V2.DefaultAPI.CreateIamRoleMapping(ctx, *r.client.OrganizationId).
		CreateIamRoleMappingRequest(tabular.CreateIamRoleMappingRequest{
			Name:       &name,
			AwsRoleArn: &awsRoleArn,
			RoleId:     &roleId,
		}).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Error creating service account", "Unable to create service account "+err.Error())
	}

	if id, ok := roleMappingAWSResponse.GetCredentialIdOk(); ok {
		plan.Id = types.StringValue(*id)
	} else {
		resp.Diagnostics.AddError("Unable to set credential_key", "Unable to set credential_key")
	}

	plan.Name = types.StringValue(name)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *awsRoleMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Service Account Update Not Supported", "A service_account update shouldn't be possible; please file an issue with the maintainers")
}

func (r *awsRoleMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state awsRoleMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	_, err := r.client.V2.DefaultAPI.DeleteServiceAccountCredential(ctx, *r.client.OrganizationId, state.Id.ValueString()).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Error deleting roleMappingAWS", "Unable to delete roleMappingAWS "+err.Error())
	}
}
