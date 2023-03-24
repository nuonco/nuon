// gke service accuont
resource "google_service_account" "gke-default-service" {
  account_id   = "gke-default-service"
  display_name = "gke-default-service"
}
