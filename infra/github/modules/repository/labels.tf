locals {
  size_labels = []

  issue_labels = [
    { name : "bug", color : "77b800", desc : "A bug." },
    { name : "improvement", color : "eb9500", desc : "A QOL improvement such as improving linting, refactoring etc." },
    { name : "auto-merge", color : "FF0000", desc : "Automatically merge this PR if it passes all PR checks." },
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
