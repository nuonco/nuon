resource "pagerduty_schedule" "schedule" {
  name = "Engineering"
  teams = [
    pagerduty_team.engineering.id,
  ]

  time_zone = "America/Los_Angeles"
  layer {
    name                         = "On-Call"
    rotation_turn_length_seconds = 345600
    rotation_virtual_start       = "2025-03-11T12:30:00-07:00"
    start                        = "2025-03-11T12:57:03-07:00"
    users = [
      data.pagerduty_user.fred.id,
      data.pagerduty_user.harsh.id,
      data.pagerduty_user.nat.id,
      data.pagerduty_user.rob.id,
      data.pagerduty_user.sam.id,
    ]
  }
}

