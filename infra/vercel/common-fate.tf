resource "vercel_project" "common_fate_installer" {
  name      = "common-fate-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/common-fate-installer"
  }
}
