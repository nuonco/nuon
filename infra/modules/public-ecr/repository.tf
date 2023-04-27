resource "aws_ecrpublic_repository" "main" {
  repository_name = var.name

  catalog_data {
    about_text      = var.about
    description     = var.description
    logo_image_blob = filebase64(var.logo_image_path)
  }

  tags = var.tags
}
