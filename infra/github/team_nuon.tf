resource "github_membership" "nuon" {
  provider = github.nuon

  for_each = local.members
  username = each.key
  role     = lookup(each.value, "role", "member")
}

resource "github_team_members" "nuonco" {
  provider = github.nuon
  team_id = github_team.nuon.id

  dynamic "members" {
    for_each = { for user, m in local.members : user => m if contains(m.teams, github_team.nuon.name) }
    content {
      username = members.key
      role     = try(members.value.role, "") == "admin" ? "maintainer" : "member"
    }
  }
}

resource "github_team" "nuon" {
  provider = github.nuon
  name        = "team"
  description = "The full Nuon team"
  privacy     = "closed"
}
