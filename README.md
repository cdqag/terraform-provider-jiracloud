# terraform-provider-jiracloud

Terraform Provider Jira Cloud

## Why

The currently available (as of 2024-07) terraform providers to JIRA Cloud are not perfect due to few reasons.
One of them is lack of the official provider (from Atlassian), but other reasons are important too, e.g. a lot of changes in JIRA CLoud REST API in previous years that lowered the API stability.
This project was started to provide a working implementation for managing JIRA CLoud Project Components bacause the existing providers did not support it fully (e.g. without possibility to set the component leads as default assignees).

## How

The go code is based on the official [Terraform Provider Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk), and uses the [`jira-go`](https://github.com/andygrunwald/go-jira) client as the interface to the [JIRA Cloud REST API](https://developer.atlassian.com/cloud/jira/platform/rest/v3/).

## Usage

### Authentication

The provider defines the following configuration options:

```hcl
variable "host" {
    description = "CDQ Jira Cloud instance url"
    type = string
}
variable "user_email" {
    description = "Email of Jira Cloud user who has access to the instance"
    type = string
}
variable "api_token" {
    description = "Jira Cloud API token of the user"
    type = string
    sensitive = true
}
```

To set the configuration variables, you can use the following environment variables:

- `JIRA_URL` for the `host` variable
- `JIRA_USER_EMAIL` for the `user_email` variable
- `JIRA_TOKEN` for the `api_token` variable

The `api_token` is the JIRA Cloud "API token", and can be generated in the Jira Cloud settings, see the [Atlassian documentation](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/) for more details.

### Datasources and Resources

The only available resource (also as data source) is `jiracloud_component`.
It has the following attributes:

- `project` (required) - the project key, e.g. `ABC`
- `name` (required) - the component name, e.g. `Component 1`
- `description` (optional) - the component description
- `assignee_type` (optional) - the assignee type, valid values are `PROJECT_DEFAULT`, `COMPONENT_LEAD`, `PROJECT_LEAD`, `UNASSIGNED`.
- `lead` (optional) - the user key of the component lead, e.g. `123120392e98rfasi47fh29ub`

Currentl version does not support deleting the components, so the `prevent_destroy` lifecycle attribute should be set to `true`:

```hcl
resource "jiracloud_component" "artifactory" {
  project = "ABC"
  name = "Component 1"
  description = "The first component of the project"
  assignee_type = "COMPONENT_LEAD"
  lead = "123120392e98rfasi47fh29ub"

  lifecycle {
    prevent_destroy = true
  }
}
```
