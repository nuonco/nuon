resource "google_kms_key_ring_iam_member" "gke-default-service" {
  key_ring_id = "us-west1/safe"
  role        = "roles/cloudkms.cryptoKeyDecrypter"
  member      = "serviceAccount:${google_service_account.gke-default-service.email}"
}
