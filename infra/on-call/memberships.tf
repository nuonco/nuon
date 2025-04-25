resource "pagerduty_team_membership" "engineering_casey_membership" {
  user_id = data.pagerduty_user.casey.id
  team_id = pagerduty_team.engineering.id
}

resource "pagerduty_team_membership" "engineering_fred_membership" {
  user_id = data.pagerduty_user.fred.id
  team_id = pagerduty_team.engineering.id
}


resource "pagerduty_team_membership" "engineering_harsh_membership" {
  user_id = data.pagerduty_user.harsh.id
  team_id = pagerduty_team.engineering.id
}


resource "pagerduty_team_membership" "engineering_jon_membership" {
  user_id = data.pagerduty_user.jon.id
  team_id = pagerduty_team.engineering.id
}


resource "pagerduty_team_membership" "engineering_jordan_membership" {
  user_id = data.pagerduty_user.jordan.id
  team_id = pagerduty_team.engineering.id
}


resource "pagerduty_team_membership" "engineering_nat_membership" {
  user_id = data.pagerduty_user.nat.id
  team_id = pagerduty_team.engineering.id
}

resource "pagerduty_team_membership" "engineering_rob_membership" {
  user_id = data.pagerduty_user.rob.id
  team_id = pagerduty_team.engineering.id
}

resource "pagerduty_team_membership" "engineering_sam_membership" {
  user_id = data.pagerduty_user.sam.id
  team_id = pagerduty_team.engineering.id
}

