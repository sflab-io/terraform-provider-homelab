package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &namingDataSource{}
)

// NewNamingDataSource is a helper function to simplify the provider implementation.
func NewNamingDataSource() datasource.DataSource {
	return &namingDataSource{}
}

// namingDataSource is the data source implementation.
type namingDataSource struct{}

// namingDataSourceModel describes the data source data model.
type namingDataSourceModel struct {
	Env  types.String `tfsdk:"env"`
	App  types.String `tfsdk:"app"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *namingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_naming"
}

// Schema defines the schema for the data source.
func (d *namingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a standardized name based on environment and application identifiers.",
		MarkdownDescription: "The `naming` data source generates standardized names following the pattern `<env>-<app>`. " +
			"This enables consistent naming conventions across all homelab infrastructure resources.",

		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				Description:         "The environment name (e.g., 'dev', 'staging', 'prod').",
				MarkdownDescription: "The environment name (e.g., `dev`, `staging`, `prod`).",
				Required:            true,
			},
			"app": schema.StringAttribute{
				Description:         "The application name (e.g., 'web', 'db', 'api').",
				MarkdownDescription: "The application name (e.g., `web`, `db`, `api`).",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Description:         "The generated name following the pattern <env>-<app>.",
				MarkdownDescription: "The generated name following the pattern `<env>-<app>`.",
				Computed:            true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *namingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data namingDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If env is "prod", do not concatenate — use the app name as-is and return early.
	if data.Env.ValueString() == "prod" || data.Env.ValueString() == "production" {
		data.Name = types.StringValue(data.App.ValueString())
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}
		return
	}

	// Generate the name by concatenating app and env with a hyphen
	name := fmt.Sprintf("%s-%s", data.App.ValueString(), data.Env.ValueString())
	data.Name = types.StringValue(name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
