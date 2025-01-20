resource "vercel_project" "datadog_installer" {
  name           = "datadog-installer"
  framework      = "nextjs"
  root_directory = "installer"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/datadog"
  }
}
