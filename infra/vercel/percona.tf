resource "vercel_project" "percona_installer" {
  name      = "percona-installer"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = "nuonco-shared/percona-installer"
  }
}
