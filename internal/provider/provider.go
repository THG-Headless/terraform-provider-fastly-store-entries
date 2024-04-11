// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &fastlystoreitemsProvider{}
var _ provider.ProviderWithFunctions = &fastlystoreitemsProvider{}

type fastlystoreitemsProvider struct {
	version string
}

type ConfiguredData struct {
	client  *http.Client
	baseUrl string
	apiKey  string
}

// ProviderModel describes the provider data model.
type ProviderModel struct {
	ApiKey  types.String `tfsdk:"api_key"`
	BaseUrl types.String `tfsdk:"base_url"`
}

func (p *fastlystoreitemsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fastlystoreitems"
	resp.Version = p.version
}

func (p *fastlystoreitemsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Fastly API Key from https://app.fastly.com/#account",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "Fastly API URL",
				Optional:    true,
			},
		},
	}
}

func (p *fastlystoreitemsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	apiToken := os.Getenv("FASTLY_API_KEY")
	baseUrl := "https://api.fastly.com"

	var data ProviderModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ApiKey.ValueString() != "" {
		apiToken = data.ApiKey.ValueString()
	}
	if data.BaseUrl.ValueString() != "" {
		apiToken = data.BaseUrl.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the FASTLY_API_KEY environment variable or provider "+
				"configuration block api_token attribute.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	client := http.DefaultClient
	var downstreamData = ConfiguredData{
		client:  client,
		baseUrl: baseUrl,
		apiKey:  apiToken,
	}
	resp.DataSourceData = &downstreamData
	resp.ResourceData = &downstreamData
}

func (p *fastlystoreitemsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewKvStoreitemResource,
	}
}

func (p *fastlystoreitemsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *fastlystoreitemsProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &fastlystoreitemsProvider{
			version: version,
		}
	}
}
