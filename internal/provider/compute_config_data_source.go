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
var _ datasource.DataSource = &ComputeConfigDataSource{}
var _ datasource.DataSourceWithConfigure = &ComputeConfigDataSource{}

func NewComputeConfigDataSource() datasource.DataSource {
	return &ComputeConfigDataSource{}
}

// ComputeConfigDataSource defines the data source implementation.
type ComputeConfigDataSource struct {
	client *util.Client
}

// ComputeConfigDataSourceModel describes the data source data model.
type ComputeConfigDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	WareHouseId   types.String `tfsdk:"warehouse_id"`
	WarehouseName types.String `tfsdk:"warehouse_name"`
	SparkConfig   types.String `tfsdk:"spark_config"`
}

func (d *ComputeConfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_config"
}

func (d *ComputeConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Tabular ComputeConfig data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform resource id",
				Computed:    true,
			},
			"warehouse_id": schema.StringAttribute{
				MarkdownDescription: "Warehouse ID",
				Optional:            true,
			},
			"warehouse_name": schema.StringAttribute{
				MarkdownDescription: "Warehouse Name",
				Optional:            true,
			},
			"spark_config": schema.StringAttribute{
				MarkdownDescription: "Spark Config that can be used to configure compute",
				Computed:            true,
			},
		},
	}
}

func (d *ComputeConfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ComputeConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Load ComputeConfigData from config
	var computeConfigData ComputeConfigDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &computeConfigData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct Warehouse Data
	var warehouseData WarehouseDataSourceModel
	warehouseData.Id = computeConfigData.WareHouseId
	warehouseData.Name = computeConfigData.WarehouseName
	GetWarehouseByIdOrName(ctx, *d.client, &warehouseData, resp)

	// Set spark config
	computeConfigData.Id = warehouseData.Id
	computeConfigData.WareHouseId = warehouseData.Id
	computeConfigData.WarehouseName = warehouseData.Name
	sparkConfig := GetIAMRoleMappingSparkConfig(warehouseData.Name.ValueString(), warehouseData.Region.ValueString())
	computeConfigData.SparkConfig = types.StringValue(sparkConfig)

	// Add ComputeConfigData to response
	resp.Diagnostics.Append(resp.State.Set(ctx, &computeConfigData)...)
}

func GetIAMRoleMappingSparkConfig(warehouseName string, warehouseRegion string) string {
	return fmt.Sprintf(`[
	  {
		"Classification": "iceberg-defaults",
		"Properties": {
		  "iceberg.enabled": "true"
		}
	  },
	  {
		"Classification": "spark-defaults",
		"Properties": {
		  "spark.sql.catalog.%[1]s": "org.apache.iceberg.spark.SparkCatalog",
		  "spark.sql.catalog.%[1]s.catalog-impl": "org.apache.iceberg.rest.RESTCatalog",
		  "spark.sql.catalog.%[1]s.rest.sigv4-enabled": "true",
		  "spark.sql.catalog.%[1]s.uri": "https://iam-gw.%[2]s.tabular.io/ws/",
		  "spark.sql.catalog.%[1]s.warehouse": "%[1]s",
		  "spark.sql.defaultCatalog": "%[1]s",
		  "spark.sql.extensions": "org.apache.iceberg.spark.extensions.IcebergSparkSessionExtensions"
		}
	  }
	]`, warehouseName, warehouseRegion)
}
