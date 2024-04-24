resource "vercel_project" "website" {
  name           = "website"
  framework      = "astro"
  root_directory = "services/website"
  ignore_command = "git diff HEAD^ HEAD --quiet -- ./"

  git_repository = {
    type = "github"
    repo = "powertoolsdev/mono"
  }
}
