resource "vercel_project" "athena_installer" {
  name      = "athena-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/athena-installer"
  }
}
