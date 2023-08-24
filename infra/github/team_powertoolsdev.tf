locals {
  # please keep alphabetical!
  members = {
    ExecutiveDre : {
      teams : [github_team.team.name, ]
    },
    jonmorehouse : {
      teams : [github_team.team.name, github_team.nuon.name]
      role : "admin"
    },
    jordan-acosta : {
      teams : [github_team.team.name, github_team.nuon.name]
      role : "admin"
    },
    nnnnat : {
      teams : [github_team.team.name, github_team.nuon.name]
      role : "admin"
    },
    nuonbot : {
      teams : [github_team.team.name, ]
    },
    pavisandhu : {
      teams : [github_team.team.name, ]
    },
    sbnoorwd : {
      teams : [github_team.team.name, ]
    },
  }
}

resource "github_membership" "powertoolsdev" {
  for_each = local.members
  username = each.key
  role     = lookup(each.value, "role", "member")
}

resource "github_team_members" "team" {
  team_id = github_team.team.id

  dynamic "members" {
    for_each = { for user, m in local.members : user => m if contains(m.teams, github_team.team.name) }
    content {
      username = members.key
      role     = try(members.value.role, "") == "admin" ? "maintainer" : "member"
    }
  }
}

resource "github_team" "team" {
  name        = "team"
  description = "The full team"
  privacy     = "closed"
}
