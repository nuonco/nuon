resource "vercel_project" "warpstream_installer" {
  name      = "warpstream-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/warpstream-installer"
  }
}
