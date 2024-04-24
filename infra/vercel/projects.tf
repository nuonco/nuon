locals {
  website_domain = "nuon.co"
}

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

resource "vercel_project_domain" "website" {
  project_id = vercel_project.website.id
  domain     = local.website_domain
}

resource "vercel_project_domain" "www-website" {
  project_id = vercel_project.website.id
  domain = "www.${local.website_domain}"

  redirect = local.website_domain
  redirect_status_code = 308
}
