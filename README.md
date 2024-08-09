# terraform-provider-jiracloud
Terraform Provider Jira Cloud

## Why

The currently available (as of 2024-07) terraform providers to JIRA Cloud are not perfect due to few reasons.
One of them is lack of the official provider (from Atlassian), but other reasons are important too, e.g. a lot of changes in JIRA CLoud REST API in previous years that lowered the API stability.
This project was started to provide a working implementation for managing JIRA CLoud Project Components bacause the existing providers did not support it fully (e.g. without possibility to set the component leads as default assignees).

## How

The go code is based on the official [Terraform Provider Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk), and uses the [`jira-go`](https://github.com/andygrunwald/go-jira) client as the interface to the [JIRA Cloud REST API](https://developer.atlassian.com/cloud/jira/platform/rest/v3/).

## Development

### Building the provider
The provider is not yet published to the Terraform Registry, so you need to build it yourself using the following steps:

1. Clone the repository
2. Run `go mod tidy` to install dependencies
3. Run `go install .` to build and install the provider binary to your `$GOPATH/bin` directory (or the `$GOBIN` directory if you have it set)

### Integrating the provider with Terraform

There are two tested and supported ways to integrate the provider with Terraform:

1. [Development Override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)
2. [Explicit Installation](https://developer.hashicorp.com/terraform/cli/config/config-file#explicit-installation-method-configuration) using `filesystem_mirror` (recommended)

#### Development Override

This method is useful for local development and testing. It allows you to use the provider without installing it to the Terraform plugin directory, and provides broader logging and debugging capabilities.

To use the "development override" method, add the following to your `~/.terraformrc` file, and replace `<GOBIN>` with the path where the built provider binary is located:

```hcl
provider_installation {
  dev_overrides {
      "cdq.com/local/jiracloud" = "<GOBIN>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

In this mode, some features of terraform are not working or are limitted, e.g. `terraform init` will not work.


#### Explicit Installation using filesystem_mirror (recommended)

This method is useful for production use (e.g. in the GitHub Actions workflow), and requires the provider to be built and installed to the Terraform plugin directory.

To use the "explicit installation" method, add the following to your `~/.terraformrc` file, and replace `<HOME>` with the path to your home directory:

```hcl
provider_installation {
  filesystem_mirror {
      path = "<HOME>/.terraform.d/plugins"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {
      exclude = ["cdq.com/*/*"]
  }
}
```

In addition, run the following commands to move the built provider binary to the plugin directory checked by Terraform (assuming that the built provider binary is located in the `$GOBIN` directory):

```sh
export TERRAFORM_JIRA_PROVIDER_NAME=terraform-provider-jiracloud
export TERRAFORM_JIRA_PROVIDER_SOURCE=cdq.com/local/jiracloud
export TERRAFORM_JIRA_PROVIDER_VERSION=1.0.0
export OS_ARCH=darwin_amd64 # or linux_amd64 or other, see sections below
export LOCAL_PROVIDER_PATH=$HOME/.terraform.d/plugins/$TERRAFORM_JIRA_PROVIDER_SOURCE/$TERRAFORM_JIRA_PROVIDER_VERSION/$OS_ARCH/
mkdir -p $LOCAL_PROVIDER_PATH
mv $GOBIN/$TERRAFORM_JIRA_PROVIDER_NAME $LOCAL_PROVIDER_PATH/$(echo $GOBIN/$TERRAFORM_JIRA_PROVIDER_NAME)_v$TERRAFORM_JIRA_PROVIDER_VERSION
```

#### Calculating the OS_ARCH

The `OS_ARCH` value is a combination of the operating system and architecture, and is used to determine the directory structure in the plugin directory. The current terraform implementation uses the go `runtime` package to determine the OS and architecture.

To calculate it to match the Terraform expected format, build and run a tool that lives in the `tools` directory of this repository:

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


