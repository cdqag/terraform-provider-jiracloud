package provider

import (
	"context"
	"os"

	jira "github.com/andygrunwald/go-jira/v2/cloud"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure JiraCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &JiraCloudProvider{}
var _ provider.ProviderWithFunctions = &JiraCloudProvider{}

// JiraCloudProvider defines the provider implementation.
type JiraCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// JiraCloudProviderModel describes the provider data model.
type JiraCloudProviderModel struct {
	Host      types.String `tfsdk:"host"`
	UserEmail types.String `tfsdk:"user_email"`
	ApiToken  types.String `tfsdk:"api_token"`
}

func (p *JiraCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "jiracloud"
	resp.Version = p.version
}

func (p *JiraCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The hostname of the Jira Cloud instance, e.g. `https://example.atlassian.net`.",
				Optional:            true,
			},
			"user_email": schema.StringAttribute{
				MarkdownDescription: "The user's email to authenticate with.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "The Jira Cloud API token to authenticate with.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *JiraCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config JiraCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Jira Cloud host",
			"The provider cannot create the Jira Cloud API client without a JIRA Cloud host url",
		)
	}

	// if config.Username.IsUnknown() {
	// 	resp.Diagnostics.AddAttributeError(
	// 		path.Root("username"),
	// 		"Unknown Jira Cloud username",
	// 		"The provider cannot create the Jira Cloud API client without a JIRA Cloud username",
	// 	)
	// }

	if config.UserEmail.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("user_email"),
			"Unknown Jira Cloud user email",
			"The provider cannot create the Jira Cloud API client without a JIRA Cloud user email",
		)
	}

	if config.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown Jira Cloud API token",
			"The provider cannot create the Jira Cloud API client without a JIRA Cloud API token",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("JIRA_URL")
	userEmail := os.Getenv("JIRA_USER_EMAIL")
	apiToken := os.Getenv("JIRA_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.UserEmail.IsNull() {
		userEmail = config.UserEmail.ValueString()
	}

	if !config.ApiToken.IsNull() {
		apiToken = config.ApiToken.ValueString()
	}

	// Check if the values are set
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Jira Cloud host",
			"The provider cannot create the Jira Cloud API client without a JIRA Cloud host url."+
				"Set the `host` attribute in the provider configuration or set the `JIRA_URL` environment variable.",
		)
	}

	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Jira Cloud API token",
			"The provider cannot create the Jira Cloud API client without a JIRA Cloud API token"+
				"Set the `api_token` attribute in the provider configuration or set the `JIRA_TOKEN` environment variable.",
		)
	}

	if userEmail == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user_email"),
			"Missing Jira Cloud user email",
			"The provider cannot create the Jira Cloud API client without a JIRA Cloud user email"+
				"Set the `user_email` attribute in the provider configuration or set the `JIRA_USER_EMAIL` environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Configure the Jira Cloud API client
	transport := jira.BasicAuthTransport{
		Username: userEmail,
		APIToken: apiToken,
	}
	httpClient := transport.Client()

	jiraClient, err := jira.NewClient(host, httpClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create Jira Cloud API client",
			"An unexpected error occurred while creating the Jira Cloud API client"+
				"Jira Cloud client error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = jiraClient
	resp.ResourceData = jiraClient
}

func (p *JiraCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewComponentResource,
	}
}

func (p *JiraCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewJiraComponentDataSource,
	}
}

func (p *JiraCloudProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &JiraCloudProvider{
			version: version,
		}
	}
}
