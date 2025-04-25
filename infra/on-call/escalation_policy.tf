resource "pagerduty_escalation_policy" "support" {
  description = null
  name        = "Support"
  num_loops   = 0
  teams = [
    pagerduty_team.engineering.id
  ]

  rule {
    escalation_delay_in_minutes = 5

    escalation_rule_assignment_strategy {
      type = "assign_to_everyone"
    }

    target {
      id   = pagerduty_user.jon.id
      type = "user_reference"
    }
  }

  rule {
    escalation_delay_in_minutes = 30

    escalation_rule_assignment_strategy {
      type = "assign_to_everyone"
    }

    target {
      id   = pagerduty_schedule.schedule.id
      type = "schedule_reference"
    }
  }
}

