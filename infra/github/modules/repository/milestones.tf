locals {
  cross_cutting_milestones = {
    "demo ready" : {
      description : "We can do a real demo for potential customers."
    },
    "customer 0" : {
      description : "We're ready to onboard the first customer."
    },
  }
}

resource "github_repository_milestone" "cross_cutting" {
  for_each = { for k, v in local.cross_cutting_milestones : k => v if !var.archived }

  owner      = split("/", github_repository.main.full_name)[0]
  repository = github_repository.main.name

  title       = each.key
  description = try(each.value.description, "")
  due_date    = try(each.value.due_date, "")
  state       = try(each.value.state, "open")
}
