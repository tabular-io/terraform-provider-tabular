package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/tabular-sdk-go/tabular"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
)

var (
	_ resource.Resource              = &storageProfileS3Resource{}
	_ resource.ResourceWithConfigure = &storageProfileS3Resource{}
)

type storageProfileS3Resource struct {
	client *util.Client
}

func NewStorageProfileS3Resource() resource.Resource {
	return &storageProfileS3Resource{}
}

type storageProfileS3ResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Region     types.String `tfsdk:"region"`
	Bucket     types.String `tfsdk:"s3_bucket_name"`
	RoleArn    types.String `tfsdk:"role_arn"`
	ExternalId types.String `tfsdk:"external_id"`
}

func (r *storageProfileS3Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*util.Client)
}

func (r *storageProfileS3Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_storage_profile"
}

func (r *storageProfileS3Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Tabular S3 Storage Profile",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Storage Profile UUID",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description:         "Storage Profile region",
				MarkdownDescription: "",
				Required:            true,
			},
			"s3_bucket_name": schema.StringAttribute{
				Description: "S3 bucket name",
				Required:    true,
			},
			"role_arn": schema.StringAttribute{
				Description: "Storage Profile IAM Role ARN",
				Required:    true,
			},
			"external_id": schema.StringAttribute{
				Description: "External ID",
				Computed:    true,
			},
		},
	}
}

func (r *storageProfileS3Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageProfileS3ResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageProfileId := state.Id.ValueString()
	storageProfile, _, err := r.client.V2.DefaultApi.GetStorageProfile(ctx, *r.client.OrganizationId, storageProfileId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error getting storage profile", "Could not get storage profile "+err.Error())
	}

	if region, ok := storageProfile.GetRegionOk(); ok {
		state.Region = types.StringValue(*region)
	} else {
		state.Region = types.StringNull()
	}

	if s3BucketName, ok := storageProfile.GetBucketOk(); ok {
		state.Bucket = types.StringValue(*s3BucketName)
	} else {
		state.Bucket = types.StringNull()
	}

	if iamRoleArn, ok := storageProfile.GetRoleArnOk(); ok {
		state.RoleArn = types.StringValue(*iamRoleArn)
	} else {
		state.RoleArn = types.StringNull()
	}

	if externalId, ok := storageProfile.GetExternalIdOk(); ok {
		state.ExternalId = types.StringValue(*externalId)
	} else {
		state.ExternalId = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageProfileS3Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageProfileS3ResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := plan.Region.ValueString()
	s3Bucket := plan.Bucket.ValueString()
	iamRoleArn := plan.RoleArn.ValueString()

	storageProfileResponse, _, err := r.client.V2.DefaultApi.CreateStorageProfile(ctx, *r.client.OrganizationId).
		CreateS3StorageProfileRequest(tabular.CreateS3StorageProfileRequest{
			Region:  &region,
			Bucket:  &s3Bucket,
			RoleArn: &iamRoleArn,
		}).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Error creating storage profile", "Unable to create storage profile "+err.Error())
		return
	}

	if id, ok := storageProfileResponse.GetIdOk(); ok {
		plan.Id = types.StringValue(*id)
	} else {
		resp.Diagnostics.AddError("Unable to set storage profile id", "Unable to set storage profile id")
	}

	if storageProfileRegion, ok := storageProfileResponse.GetRegionOk(); ok {
		plan.Region = types.StringValue(*storageProfileRegion)
	} else {
		plan.Region = types.StringNull()
	}

	if s3BucketName, ok := storageProfileResponse.GetBucketOk(); ok {
		plan.Bucket = types.StringValue(*s3BucketName)
	} else {
		plan.Bucket = types.StringNull()
	}

	if storageProfileIAMRoleArn, ok := storageProfileResponse.GetRoleArnOk(); ok {
		plan.RoleArn = types.StringValue(*storageProfileIAMRoleArn)
	} else {
		plan.RoleArn = types.StringNull()
	}

	if externalId, ok := storageProfileResponse.GetExternalIdOk(); ok {
		plan.ExternalId = types.StringValue(*externalId)
	} else {
		resp.Diagnostics.AddError("Unable to set external id", "Unable to set external id")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *storageProfileS3Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Database Update Not Supported", "A database update shouldn't be possible; please file an issue with the maintainers")
}

func (r *storageProfileS3Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageProfileS3ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	storageProfileId := state.Id.ValueString()
	_, err := r.client.V2.DefaultApi.DeleteStorageProfile(ctx, *r.client.OrganizationId, storageProfileId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting storage profile", "Unable to delete storage profile "+err.Error())
	}
}
