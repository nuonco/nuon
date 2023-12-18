resource "github_membership" "nuonco-shared" {
  provider = github.nuonco-shared

  for_each = { for _, user in local.vars.members : user.username => lookup(user, "role", "member") }

  username = each.key
  role     = each.value
}

resource "github_team_members" "nuonco-shared" {
  provider = github.nuonco-shared
  team_id  = github_team.nuonco-shared.id

  dynamic "members" {
    for_each = { for _, user in local.vars.members : user.username => lookup(user, "role", "member") }

    content {
      username = members.key
      role     = members.value == "admin" ? "maintainer" : "member"
    }
  }
}

resource "github_team" "nuonco-shared" {
  provider    = github.nuonco-shared
  name        = "team"
  description = "The full Nuon team"
  privacy     = "closed"
}
