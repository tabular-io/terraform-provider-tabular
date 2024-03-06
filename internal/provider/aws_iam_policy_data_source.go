package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tabular-io/terraform-provider-tabular/internal/provider/util"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AWSIAMPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &AWSIAMPolicyDataSource{}

func NewAWSIAMPolicyDataSource() datasource.DataSource {
	return &AWSIAMPolicyDataSource{}
}

// AWSIAMPolicyDataSource defines the data source implementation.
type AWSIAMPolicyDataSource struct {
	client *util.Client
}

// AWSIAMPolicyDataSourceModel describes the data source data model.
type AWSIAMPolicyDataSourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Bucket             types.String `tfsdk:"bucket"`
	IAMReadWritePolicy types.String `tfsdk:"iam_read_write_policy"`
	IAMReadOnlyPolicy  types.String `tfsdk:"iam_read_only_policy"`
	AssumeRolePolicy   types.String `tfsdk:"assume_role_policy"`
}

func (d *AWSIAMPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_iam_policy"
}

func (d *AWSIAMPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Tabular AWS IAMPolicy data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform resource id",
				Computed:    true,
			},
			"bucket": schema.StringAttribute{
				MarkdownDescription: "The storage bucket",
				Required:            true,
			},
			"iam_read_write_policy": schema.StringAttribute{
				MarkdownDescription: "IAM Read Write Policy",
				Computed:            true,
			},
			"iam_read_only_policy": schema.StringAttribute{
				MarkdownDescription: "IAM Read Only Policy",
				Computed:            true,
			},
			"assume_role_policy": schema.StringAttribute{
				MarkdownDescription: "Assume Role Policy",
				Computed:            true,
			},
		},
	}
}

func (d *AWSIAMPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func IAMReadWritePolicy(bucket string) string {
	return fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Action": [
			  "s3:ListBucket",
			  "s3:GetBucketLocation",
			  "s3:GetBucketNotification",
			  "s3:PutBucketNotification"
			],
			"Resource": [
			  "arn:aws:s3:::%[1]v"
			]
		  },
		  {
			"Effect": "Allow",
			"Action": [
			  "s3:PutObject",
			  "s3:GetObject",
			  "s3:DeleteObject",
			  "s3:PutObjectAcl",
			  "s3:AbortMultipartUpload"
			],
			"Resource": [
			  "arn:aws:s3:::%[1]v/*"
			]
		  }
		]
	  }`, bucket)
}

func IAMReadOnlyPolicy(bucket string) string {
	return fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Action": [
			  "s3:ListBucket",
			  "s3:GetBucketLocation",
			  "s3:GetBucketNotification"
			],
			"Resource": [
			  "arn:aws:s3:::%[1]v"
			]
		  },
		  {
			"Effect": "Allow",
			"Action": [
			  "s3:GetObject"
			],
			"Resource": [
			  "arn:aws:s3:::%[1]v/*"
			]
		  }
		]
	  }`, bucket)
}

func AssumeRolePolicy(externalId string) string {
	return fmt.Sprintf(`{
		"Version": "2008-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": "arn:aws:iam::237881912361:root"
				},
				"Action": [
					"sts:AssumeRole",
					"sts:TagSession"
				],
				"Condition": {
					"StringEquals": {
						"sts:ExternalId": "%s"
					},
					"ArnLike": {
						"aws:PrincipalArn": "arn:aws:iam::237881912361:role/TabularSignerServiceRole*"
					}
				}
			}
		]
	}`, externalId)
}

func (d *AWSIAMPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AWSIAMPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = data.Bucket
	data.IAMReadWritePolicy = types.StringValue(IAMReadWritePolicy(data.Bucket.ValueString()))
	data.IAMReadOnlyPolicy = types.StringValue(IAMReadOnlyPolicy(data.Bucket.ValueString()))
	data.AssumeRolePolicy = types.StringValue(AssumeRolePolicy(*d.client.OrganizationId))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
