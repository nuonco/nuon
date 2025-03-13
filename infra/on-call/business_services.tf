resource "pagerduty_business_service" "controlplane" {
  name             = "Control Plane API0"
  team             = "PG5IHQW"
}


resource "pagerduty_business_service" "vendor-dashbaord" {
  name             = "Vendor dashboard0"
  team             = "PG5IHQW"
}


resource "pagerduty_business_service" "installers" {
  name             = "Installers0"
  team             = "PG5IHQW"
}


resource "pagerduty_business_service" "website" {
  name             = "website0"
  team             = "PG5IHQW"
}


resource "pagerduty_business_service" "runners" {
  name             = "Runners0"
  team             = "PG5IHQW"
}
