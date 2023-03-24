// infrastructure for our build infrastructure
resource "google_storage_bucket" "build-manifests" {
  name     = "powertools-build-manifests"
  location = var.region

  versioning {
    enabled = true
  }

  lifecycle {
    prevent_destroy = false
  }
}

// infrastructure for our build infrastructure
resource "google_storage_bucket" "growth-data" {
  name     = "powertools-growth-data"
  location = var.region

  versioning {
    enabled = true
  }

  lifecycle {
    prevent_destroy = false
  }
}
