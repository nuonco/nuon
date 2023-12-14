resource "github_membership" "nuon" {
  provider = github.nuon

  for_each = { for _, user in local.vars.members : user.username => lookup(user, "role", "member") }

  username = each.key
  role     = each.value
}

resource "github_team_members" "nuon" {
  provider = github.nuon
  team_id  = github_team.nuon.id

  dynamic "members" {
    for_each = { for _, user in local.vars.members : user.username => lookup(user, "role", "member") }

    content {
      username = members.key
      role     = members.value == "admin" ? "maintainer" : "member"
    }
  }
}

resource "github_team" "nuon" {
  provider    = github.nuon
  name        = "team"
  description = "The full Nuon team"
  privacy     = "closed"
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
