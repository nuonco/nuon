resource "vercel_project" "demo_installer" {
  name      = "demo-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco/demo-installer"
  }
}
