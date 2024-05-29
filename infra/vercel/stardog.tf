resource "vercel_project" "stardog_installer" {
  name      = "stardog-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/stardog-installer"
  }
}
