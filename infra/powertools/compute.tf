#resource "google_container_cluster" "main" {
#name = "main"
#location = var.region

## We can't create a cluster with no node pool defined, but we want to only use
## separately managed node pools. So we create the smallest possible default
## node pool and immediately delete it.
#remove_default_node_pool = false
#initial_node_count = 1

#master_auth {
#username = ""
#password = ""
#}

#node_config {
#preemptible = true
#oauth_scopes = [
#"https://www.googleapis.com/auth/compute",
#"https://www.googleapis.com/auth/devstorage.read_only",
#"https://www.googleapis.com/auth/logging.write",
#"https://www.googleapis.com/auth/monitoring",
#]

#labels = {
#default_node_pool = "true"
#}
#tags = []
#}

#timeouts {
#create = "30m"
#update = "40m"
#}
#}

#resource "google_container_node_pool" "primary_preemptible_nodes" {
#name = "main-node-pool"
#location = var.region
#cluster = google_container_cluster.main.name

#node_count = 1

#node_config {
#preemptible  = true
#machine_type = "n1-standard-1"

#oauth_scopes = [
#"https://www.googleapis.com/auth/compute",
#"https://www.googleapis.com/auth/devstorage.read_only",
#"https://www.googleapis.com/auth/logging.write",
#"https://www.googleapis.com/auth/monitoring",
#]
#}
#}
