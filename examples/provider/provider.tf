terraform {
  required_providers {
    jiracloud = {
      source  = "cdqag/jiracloud"
      version = "~> 1.0.0"
    }
  }
}

provider "jiracloud" {
  host = var.host
  user_email = var.user_email
  api_token = var.api_token
}
