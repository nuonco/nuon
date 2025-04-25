resource "pagerduty_business_service" "controlplane" {
  name = "Control Plane API"
  team = "PG5IHQW"
}


resource "pagerduty_business_service" "vendor-dashboard" {
  name = "Vendor dashboard"
  team = "PG5IHQW"
}


resource "pagerduty_business_service" "installers" {
  name = "Installers"
  team = "PG5IHQW"
}


resource "pagerduty_business_service" "website" {
  name = "website"
  team = "PG5IHQW"
}


resource "pagerduty_business_service" "runners" {
  name = "Runners"
  team = "PG5IHQW"
}
