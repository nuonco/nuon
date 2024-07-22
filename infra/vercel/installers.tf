resource "vercel_project" "installers" {
  name      = "installers"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/installer-hosted"
  }
}
