resource "pagerduty_service" "ctl-api" {
    name                    = "ctl-api-0"   
    escalation_policy       = pagerduty_escalation_policy.support.id
}

resource "pagerduty_service" "dashboard-ui" {
    name                    = "dashboard-ui-0"   
    escalation_policy       = pagerduty_escalation_policy.support.id
}

resource "pagerduty_service" "installer" {
    name                    = "installers-0"   
    escalation_policy       = pagerduty_escalation_policy.support.id
}


resource "pagerduty_service" "runner" {
    name                    = "runner-0"   
    escalation_policy       = pagerduty_escalation_policy.support.id
}

resource "pagerduty_service" "terraform-cloud" {
    name                    = "terraform-cloud-0"   
    escalation_policy       = pagerduty_escalation_policy.support.id
}


resource "pagerduty_service" "website" {
    name                    = "website-0"   
    escalation_policy       = pagerduty_escalation_policy.support.id
}

resource "pagerduty_service" "workers-executors" {
    name                    = "workers-executors-0"   
    escalation_policy       = pagerduty_escalation_policy.support.id
}