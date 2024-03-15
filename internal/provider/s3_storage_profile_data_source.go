package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
)

var _ datasource.DataSource = &S3StorageProfileDataSource{}
var _ datasource.DataSourceWithConfigure = &S3StorageProfileDataSource{}

func NewS3StorageProfileDataSource() datasource.DataSource {
	return &S3StorageProfileDataSource{}
}

type S3StorageProfileDataSource struct {
	client *util.Client
}

type S3StorageProfileDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	AccountId      types.String `tfsdk:"account_id"`
	Region         types.String `tfsdk:"region"`
	Name           types.String `tfsdk:"name"`
	RoleArn        types.String `tfsdk:"role_arn"`
	ExternalId     types.String `tfsdk:"external_id"`
}

func (d *S3StorageProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_storage_profile"
}

func (d *S3StorageProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "S3StorageProfile data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "S3StorageProfile ID",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Tabular Organization ID",
				Computed:    true,
			},
			"account_id": schema.StringAttribute{
				Description: "Storage Profile AWS Account ID",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Storage Profile region",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Storage Profile bucket name",
				Required:    true,
			},
			"role_arn": schema.StringAttribute{
				Description: "Storage Profile AWS Role Arn",
				Computed:    true,
			},
			"external_id": schema.StringAttribute{
				Description: "External ID",
				Computed:    true,
			},
		},
	}
}

func (d *S3StorageProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*util.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *S3StorageProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data S3StorageProfileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	storageProfile, _, err := d.client.V2.DefaultAPI.GetStorageProfile(ctx, *d.client.OrganizationId, name).Type_("name").Execute()
	if err != nil {
		resp.Diagnostics.AddError("S3 Storage Profile not found", err.Error())
		return
	}

	if id, ok := storageProfile.GetIdOk(); ok {
		data.Id = types.StringValue(*id)
	} else {
		data.Id = types.StringNull()
	}

	if organizationId, ok := storageProfile.GetOrganizationIdOk(); ok {
		data.OrganizationId = types.StringValue(*organizationId)
	} else {
		data.OrganizationId = types.StringNull()
	}

	if accountId, ok := storageProfile.GetAccountIdOk(); ok {
		data.AccountId = types.StringValue(*accountId)
	} else {
		data.AccountId = types.StringNull()
	}

	if region, ok := storageProfile.GetRegionOk(); ok {
		data.Region = types.StringValue(*region)
	} else {
		data.Region = types.StringNull()
	}

	if organizationId, ok := storageProfile.GetOrganizationIdOk(); ok {
		data.OrganizationId = types.StringValue(*organizationId)
	} else {
		data.OrganizationId = types.StringNull()
	}

	if roleArn, ok := storageProfile.GetRoleArnOk(); ok {
		data.RoleArn = types.StringValue(*roleArn)
	} else {
		data.RoleArn = types.StringNull()
	}

	if externalId, ok := storageProfile.GetExternalIdOk(); ok {
		data.ExternalId = types.StringValue(*externalId)
	} else {
		data.ExternalId = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
