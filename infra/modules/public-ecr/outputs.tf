output "repository_url" {
  value = aws_ecrpublic_repository.main.repository_uri
}

output "repository_arn" {
  value = aws_ecrpublic_repository.main.arn
}

output "registry_id" {
  value = aws_ecrpublic_repository.main.registry_id
}

output "registry_url" {
  value = "${aws_ecrpublic_repository.main.registry_id}.dkr.ecr.${var.region}.amazonaws.com"
}

output "region" {
  value = var.region
}

output "is_public" {
  value = true
}

output "all" {
  description = "mapping output of all values"

  value = {
    repository_url = aws_ecrpublic_repository.main.repository_uri
    registry_id    = aws_ecrpublic_repository.main.registry_id
    registry_url   = "${aws_ecrpublic_repository.main.registry_id}.dkr.ecr.${var.region}.amazonaws.com"
    repository_arn = aws_ecrpublic_repository.main.arn
    is_public      = true
    region         = var.region
  }
}
