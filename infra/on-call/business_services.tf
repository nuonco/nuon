resource "pagerduty_business_service" "controlplane" {
  name = "Control Plane API"
  team = pagerduty_team.engineering.id
}


resource "pagerduty_business_service" "vendor-dashboard" {
  name = "Vendor dashboard"
  team = pagerduty_team.engineering.id
}


resource "pagerduty_business_service" "installers" {
  name = "Installers"
  team = pagerduty_team.engineering.id
}


resource "pagerduty_business_service" "website" {
  name = "website"
  team = pagerduty_team.engineering.id
}


resource "pagerduty_business_service" "runners" {
  name = "Runners"
  team = pagerduty_team.engineering.id
}
