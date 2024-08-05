package provider

import (
	"context"
	"fmt"
	"net/http"

	jira "github.com/andygrunwald/go-jira/v2/cloud"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &ComponentResource{}
	_ resource.ResourceWithImportState = &ComponentResource{}
)

func NewComponentResource() resource.Resource {
	return &ComponentResource{}
}

// ComponentResource defines the resource implementation.
type ComponentResource struct {
	client *jira.Client
}

func (r *ComponentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*jira.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *jira.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

type JiraComponentResourceModel struct {
	Project      types.String `tfsdk:"project"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	AssigneeType types.String `tfsdk:"assignee_type"`
	Lead         types.String `tfsdk:"lead"`
}

func (r *ComponentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

func (r *ComponentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"assignee_type": schema.StringAttribute{
				MarkdownDescription: "The assignee type of the Jira component." +
					"Valid values are `PROJECT_DEFAULT`, `COMPONENT_LEAD`, `PROJECT_LEAD`, `UNASSIGNED`.",
				Optional: true,
				Default:  stringdefault.StaticString("PROJECT_DEFAULT"),
				Computed: true,
			},
			"lead": schema.StringAttribute{
				MarkdownDescription: "The lead of the Jira component represented by their Jira account ID.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *ComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state JiraComponentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	options := jira.ComponentCreateOptions{
		Name:          state.Name.ValueString(),
		Description:   state.Description.ValueString(),
		LeadAccountId: state.Lead.ValueString(),
		Project:       state.Project.ValueString(),
		AssigneeType:  state.AssigneeType.ValueString(),
	}

	newComponent, _, err := r.client.Component.Create(context.Background(), &options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create component",
			fmt.Sprintf("An unexpected error occurred while creating a new component named %s... ", state.Name.ValueString())+
				"Jira Cloud client error: "+err.Error(),
		)
		return
	}

	state = JiraComponentResourceModel{
		Project:      types.StringValue(newComponent.Project),
		Name:         types.StringValue(newComponent.Name),
		Description:  types.StringValue(newComponent.Description),
		AssigneeType: types.StringValue(newComponent.AssigneeType),
		Lead:         types.StringValue(newComponent.Lead.AccountID),
	}

	tflog.Trace(ctx, fmt.Sprintf("created a brand new component (ID: %s)", newComponent.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state JiraComponentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	project, _, err := r.client.Project.Get(context.Background(), state.Project.ValueString())
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

	options := jira.ComponentCreateOptions{
		Name:          state.Name.ValueString(),
		Description:   state.Description.ValueString(),
		LeadAccountId: state.Lead.ValueString(),
		Project:       state.Project.ValueString(),
		AssigneeType:  state.AssigneeType.ValueString(),
	}

	apiEndpoint := fmt.Sprintf("rest/api/3/component/%s", projectComponentSimple.ID)
	lowLevelRequestToJiraAPI, err := r.client.NewRequest(context.Background(), http.MethodPut, apiEndpoint, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update component",
			"An unexpected error occurred while preparing a low level request to Jira API to update an existing project component... "+
				"Error: "+err.Error(),
		)
		return
	}

	updatedComponent := new(jira.ProjectComponent)
	_, err = r.client.Do(lowLevelRequestToJiraAPI, updatedComponent)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update component",
			"An unexpected error occurred while updating an existing project component... "+
				"Error: "+err.Error(),
		)
		return
	}

	state = JiraComponentResourceModel{
		Name:        types.StringValue(updatedComponent.Name),
		Description: types.StringValue(updatedComponent.Description),
		Lead:        types.StringValue(updatedComponent.Lead.AccountID),
	}

	tflog.Trace(ctx, fmt.Sprintf("created a brand new component (ID: %s)", updatedComponent.ID))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state JiraComponentResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	project, _, err := r.client.Project.Get(context.Background(), state.Project.ValueString())
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

	projectComponentEnriched, _, err := r.client.Component.Get(context.Background(), projectComponentSimple.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read component",
			fmt.Sprintf("An unexpected error occurred while reading the \"%s\" component", state.Name.ValueString())+
				"Jira Cloud client error: "+err.Error(),
		)
		return
	}

	state = JiraComponentResourceModel{
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

func (r *ComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Delete Not Implemented", "This resource does not support deletion.")
}

func (r *ComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
