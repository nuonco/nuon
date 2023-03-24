output "root_domain_nameservers" {
  value = join(",", data.google_dns_managed_zone.root.name_servers)
}

// beta gke creds
#output "main_cluster_client_certificate" {
#value = google_container_cluster.main.master_auth.0.client_certificate
#}

#output "main_cluster_client_key" {
#value = google_container_cluster.main.master_auth.0.client_key
#}

#output "main_cluster_ca_certificate" {
#value = google_container_cluster.main.master_auth.0.cluster_ca_certificate
#}

output "last_update" {
  description = "The timestamp of when the module was last applied. Useful for forcing applies on upgrade"
  value       = timestamp()
}
