package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &homelabProvider{}
)

// New is a helper function to simplify provider server setup.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &homelabProvider{
			version: version,
		}
	}
}

// homelabProvider is the provider implementation.
type homelabProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *homelabProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "homelab"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *homelabProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The homelab provider enables standardized resource naming for homelab infrastructure.",
	}
}

// Configure prepares a provider with the configuration data.
// Since this provider doesn't require any configuration or shared clients,
// this method is a no-op.
func (p *homelabProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// No provider-level configuration needed for this simple provider
}

// DataSources defines the data sources implemented in the provider.
func (p *homelabProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNamingDataSource,
	}
}

// Resources defines the resources implemented in the provider.
// This provider doesn't implement any resources.
func (p *homelabProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}
