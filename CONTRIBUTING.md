# Contributing

The go code is based on the official [Terraform Provider Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk), and uses the [`go-github`](https://github.com/google/go-github) client as the interface to the [GitHub REST API](https://docs.github.com/en/rest?apiVersion=2022-11-28).

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
