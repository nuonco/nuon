output "repository_url" {
  value = aws_ecrpublic_repository.main.repository_uri
}

output "repository_arn" {
  value = aws_ecrpublic_repository.main.arn
}

output "registry_id" {
  value = aws_ecrpublic_repository.main.registry_id
}

output "is_public" {
  value = true
}
