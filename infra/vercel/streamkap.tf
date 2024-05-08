resource "vercel_project" "streamkap_installer" {
  name      = "streamkap-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/streamkap-installer"
  }
}
