resource "github_repository_collaborator" "collaborators" {
  for_each = var.collaborators

  repository = github_repository.main.name
  username = each.key
  permission = each.value
}
