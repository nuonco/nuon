output "buildkit_namespace" {
  description = "Namespace where buildkit is deployed"
  value       = local.vars.namespace
}

output "buildkit_service" {
  description = "Service name for buildkit"
  value       = "buildkitd.${local.vars.namespace}.svc.cluster.local"
}

output "buildkit_port" {
  description = "Port for buildkit service"
  value       = local.vars.buildkit.port
}

output "proxy_service" {
  description = "Service name for buildkit proxy"
  value       = "proxy.${local.vars.namespace}.svc.cluster.local"
}
