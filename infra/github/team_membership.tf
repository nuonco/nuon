locals {
  # please keep alphabetical!
  members = {
    alphavibe : {
      teams : [github_team.team.name, ]
    },
    chrispotter : {
      teams : [github_team.team.name, github_team.backend.name, ]
    },
    ExecutiveDre : {
      teams : [github_team.team.name, ]
    },
    focusaurus : {
      teams : [github_team.team.name, github_team.backend.name, ]
    },
    jonmorehouse : {
      teams : [github_team.team.name, github_team.backend.name, github_team.frontend.name, ]
      role : "admin"
    },
    mpaktiti : {
      teams : [github_team.team.name, github_team.backend.name, ]
    },
    nnnnat : {
      teams : [github_team.team.name, github_team.frontend.name, ]
    },
    nuonbot : {
      teams : [github_team.team.name, github_team.frontend.name, ]
    },
    NuonSara : {
      teams : [github_team.team.name, github_team.frontend.name, ]
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

resource "github_team_members" "backend" {
  team_id = github_team.backend.id

  dynamic "members" {
    for_each = { for user, m in local.members : user => m if contains(m.teams, github_team.backend.name) }
    content {
      username = members.key
      role     = try(members.value.role, "") == "admin" ? "maintainer" : "member"
    }
  }
}

resource "github_team_members" "frontend" {
  team_id = github_team.frontend.id

  dynamic "members" {
    for_each = { for user, m in local.members : user => m if contains(m.teams, github_team.frontend.name) }
    content {
      username = members.key
      role     = try(members.value.role, "") == "admin" ? "maintainer" : "member"
    }
  }
}
