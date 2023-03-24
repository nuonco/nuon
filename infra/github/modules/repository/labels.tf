locals {
  size_labels = [
    { name : "size/XS", color : "00f000", lines : "0-9" },
    { name : "size/S", color : "77b800", lines : "10-29" },
    { name : "size/M", color : "ebb800", lines : "30-99" },
    { name : "size/L", color : "eb9500", lines : "100-499" },
    { name : "size/XL", color : "ff823f", lines : "500-999" },
    { name : "size/XXL", color : "ee0000", lines : "1000+" },
  ]

  issue_labels = [
    { name : "proposal", color : "00f000", desc : "A proposal for an improvement, feature request or other." },
    { name : "bug", color : "77b800", desc : "A bug." },
    { name : "improvement", color : "eb9500", desc : "A QOL improvement such as improving linting, refactoring etc." },
    { name : "product", color : "ebb800", desc : "Part of the direct product (api, ui, etc)." },
    { name : "infra", color : "eb9500", desc : "Infrastructure related to our product (sandboxes, waypoint, etc)." },
    { name : "internal-infra", color : "77b800", desc : "Internal infra related to our own (CI, AWS, workflows etc)." },
    { name : "blog-content", color : "ffa500", desc : "Potential blog content." },
  ]
}

# We have an action to update PR labels with size info, these are the standards
resource "github_issue_label" "size_labels" {
  for_each = { for _, label in local.size_labels : label.name => label }

  repository  = github_repository.main.name
  name        = each.value.name
  color       = each.value.color
  description = "Denotes a PR that changes ${each.value.lines} lines, ignoring generated files."
}

# We configure issue labels to organize and visualize issues on a per-repo basis, in addition to our cross repo views
# from projects.
resource "github_issue_label" "issue_labels" {
  for_each = { for _, label in local.issue_labels : label.name => label if !var.archived }

  repository  = github_repository.main.name
  name        = each.value.name
  color       = each.value.color
  description = each.value.desc
}
