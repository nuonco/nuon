resource "vercel_project" "weaviate_installer" {
  name      = "weaviate-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/weaviate-installer"
  }
}
