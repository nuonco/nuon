import {
  to = pagerduty_user.casey
  id = "PYBJK0N"
}

resource "pagerduty_user" "casey" {
  name  = "Casey"
  email = "casey@nuon.co"
  role  = "admin"
}

data "pagerduty_user" "casey" {
  email = "casey@nuon.co"
}

data "pagerduty_user" "fred" {
  email = "fred@nuon.co"
}


data "pagerduty_user" "harsh" {
  email = "harsh@nuon.co"
}


data "pagerduty_user" "jon" {
  email = "jon@nuon.co"
}


data "pagerduty_user" "jordan" {
  email = "jordan@nuon.co"
}

data "pagerduty_user" "nat" {
  email = "nat@nuon.co"
}


data "pagerduty_user" "rob" {
  email = "rob@nuon.co"
}


data "pagerduty_user" "sam" {
  email = "sam@nuon.co"
}


data "pagerduty_user" "tim" {
  email = "tim@nuon.co"
}

