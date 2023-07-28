output "app" {
  value = data.nuon_app.cool_app
}

output "org" {
  value = data.nuon_org.org
}

output "mono" {
  value = data.nuon_connected_repo.mono
}

output "all_repos" {
  value = data.nuon_connected_repos.all
}
