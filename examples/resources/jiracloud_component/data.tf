resource "jiracloud_component" "referencedata-fetch-and-transform-job" {
  project = "ABC"
  name = "My Component"
  description = "My Super Duper Component"
  assignee_type = "COMPONENT_LEAD"
  lead = "1a2b3c4d5e6f"

  lifecycle {
    prevent_destroy = true
  }
}
