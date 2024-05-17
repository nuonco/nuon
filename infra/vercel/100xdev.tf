resource "vercel_project" "_100xdev_installer" {
  name      = "100xdev-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/100xdev-installer"
  }
}
