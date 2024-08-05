package provider

import (
	"context"
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &JiraComponentDataSource{}
	_ datasource.DataSourceWithConfigure = &JiraComponentDataSource{}
)

func NewJiraComponentDataSource() datasource.DataSource {
	return &JiraComponentDataSource{}
}

// JiraComponentDataSource defines the data source implementation.
type JiraComponentDataSource struct {
	client *jira.Client
}

func (d *JiraComponentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*jira.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *jira.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

type JiraComponentDataSourceModel struct {
	Project     types.String `tfsdk:"project"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Lead        types.String `tfsdk:"lead"`
}

func (d *JiraComponentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

func (d *JiraComponentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Jira Component Data Source",

		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				MarkdownDescription: "The Jira project key that the component belongs to.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Jira component.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the Jira component.",
				Optional:            true,
				Computed:            true,
			},
			"lead": schema.StringAttribute{
				MarkdownDescription: "The lead of the Jira component represented by their Jira account ID.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (d *JiraComponentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state JiraComponentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	project, _, err := d.client.Project.Get(context.Background(), state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to read %s project", state.Project.ValueString()),
			fmt.Sprintf("An unexpected error occurred while reading the %s project... ", state.Project.ValueString())+
				"Jira Cloud client error: "+err.Error(),
		)
		return
	}

	var projectComponentSimple jira.ProjectComponent
	for _, component := range project.Components {
		if component.Name == state.Name.ValueString() {
			projectComponentSimple = component
			break
		}
	}

	if projectComponentSimple.ID == "" {
		resp.Diagnostics.AddError(
			"Failed to find component",
			"Could not find a component with the name: "+state.Name.String(),
		)
		return
	}

	projectComponentEnriched, _, err := d.client.Component.Get(context.Background(), projectComponentSimple.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read component",
			fmt.Sprintf("An unexpected error occurred while reading the \"%s\" component", state.Name.ValueString())+
				"Jira Cloud client error: "+err.Error(),
		)
		return
	}

	state = JiraComponentDataSourceModel{
		Project:     types.StringValue(project.Key),
		Name:        types.StringValue(projectComponentEnriched.Name),
		Description: types.StringValue(projectComponentEnriched.Description),
		Lead:        types.StringValue(projectComponentEnriched.Lead.AccountID),
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
