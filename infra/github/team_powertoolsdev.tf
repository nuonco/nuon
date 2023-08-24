resource "github_membership" "powertoolsdev" {
  for_each = { for _, user in local.vars.members : user.username => lookup(user, "role", "member") }

  username = each.key
  role     = each.value
}

resource "github_team_members" "team" {
  team_id = github_team.team.id

  dynamic "members" {
    for_each = { for _, user in local.vars.members : user.username => lookup(user, "role", "member") }

    content {
      username = members.key
      role     = members.value == "admin" ? "maintainer" : "member"
    }
  }
}

resource "github_team" "team" {
  name        = "team"
  description = "The full team"
  privacy     = "closed"
}
