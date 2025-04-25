resource "pagerduty_service" "ctl-api" {
  name              = "ctl-api"
  escalation_policy = pagerduty_escalation_policy.support.id
}

resource "pagerduty_service" "dashboard-ui" {
  name              = "dashboard-ui"
  escalation_policy = pagerduty_escalation_policy.support.id
}

resource "pagerduty_service" "installer" {
  name              = "installers"
  escalation_policy = pagerduty_escalation_policy.support.id
}


resource "pagerduty_service" "runner" {
  name              = "runner"
  escalation_policy = pagerduty_escalation_policy.support.id
}

resource "pagerduty_service" "terraform-cloud" {
  name              = "terraform-cloud"
  escalation_policy = pagerduty_escalation_policy.support.id
}


resource "pagerduty_service" "website" {
  name              = "website"
  escalation_policy = pagerduty_escalation_policy.support.id
}

resource "pagerduty_service" "workers-executors" {
  name              = "workers-executors"
  escalation_policy = pagerduty_escalation_policy.support.id
}


resource "pagerduty_service" "unrouted-events" {
  name              = "unrouted-events"
  escalation_policy = pagerduty_escalation_policy.support.id
}

