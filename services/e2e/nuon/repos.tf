data "nuon_connected_repo" "mono" {
  org_id = data.nuon_org.org.id
  name = "powertoolsdev/mono"
}

data "nuon_connected_repos" "all" {
  org_id = data.nuon_org.org.id
}
