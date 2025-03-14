resource "pagerduty_schedule" "schedule" {
    name           = "Engineering"
    teams          = [
       pagerduty_team.engineering.id,
    ]

    time_zone      = "America/Los_Angeles"
    layer {
        name                         = "On-Call-0"
        rotation_turn_length_seconds = 345600
        rotation_virtual_start       = "2025-03-11T12:30:00-07:00"
        start                        = "2025-03-11T12:57:03-07:00"
        users                        = [
            pagerduty_user.fred.id,
            pagerduty_user.harsh.id,
            pagerduty_user.jon.id,
            pagerduty_user.nat.id,
            pagerduty_user.rob.id,
            pagerduty_user.sam.id,
        ]
    }
}