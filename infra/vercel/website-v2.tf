resource "vercel_project" "website_v2" {
  name           = "website-v2"
  framework      = "astro"
  root_directory = "services/website-v2"
  ignore_command = "git diff HEAD^ HEAD --quiet -- ./"

  git_repository = {
    type = "github"
    repo = "powertoolsdev/mono"
  }
}
