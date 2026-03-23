output "snapshot_id" {
  description = "ID of the EBS snapshot containing pre-cached container images"
  value       = aws_ebs_snapshot.cache.id
}

output "snapshot_arn" {
  description = "ARN of the EBS snapshot"
  value       = aws_ebs_snapshot.cache.arn
}

output "volume_size_gb" {
  description = "Size of the cache volume in GB (use this for blockDeviceMappings volumeSize)"
  value       = var.volume_size_gb
}
