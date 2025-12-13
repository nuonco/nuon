locals {
  nuon_labs_domain = "labs.nuon.co"
}

resource "vercel_project" "nuon_labs" {
  name           = "nuon-labs"
  framework      = "nextjs"
  root_directory = "services/nuon-labs"
  ignore_command = "git diff HEAD^ HEAD --quiet -- ./"

  git_repository = {
    type = "github"
    repo = "powertoolsdev/mono"
  }
}

resource "vercel_project_domain" "nuon_labs" {
  project_id = vercel_project.nuon_labs.id
  domain     = local.nuon_labs_domain
}
