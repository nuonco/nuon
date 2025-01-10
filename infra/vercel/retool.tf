resource "vercel_project" "retool_installer" {
  name           = "retool-installer"
  framework      = "nextjs"
  root_directory = "installer"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/retool-installer"
  }
}
