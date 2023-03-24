resource "github_team" "team" {
  name        = "team"
  description = "The full Nuon team"
  privacy     = "closed"
}

resource "github_team" "backend" {
  name           = "backend"
  description    = "Backend engineers"
  privacy        = "closed"
  parent_team_id = github_team.team.id
}

resource "github_team" "frontend" {
  name           = "frontend"
  description    = "Frontend engineers"
  privacy        = "closed"
  parent_team_id = github_team.team.id
}

# NOTE(jdt): when adding a team that has a parent, an admin will have to accept the invite
