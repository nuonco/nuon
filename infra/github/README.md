# infra-github

Github Configuration as Code

## About

This repository includes repository configuration for all nuon repositories.
Currently, repos are in the `powertoolsdev` org.

We've configured the github provider to use app credentials instead of long-
lived, static creds. 

## Lifecycle

### Adding a new repo

1. In `repositories.tf`, add a module block in the correct spot alphabetically.
1. Name the resource the same as the desired repo name and fill out the desired
inputs.
1. `terraform init && terraform plan`
1. Create PR, review plan, merge, approve plan.

### Archiving a repo

1. In `repositories.tf`, find the resource representing the repo to be archived.
1. Add `archived = true`
1. Add `"archived"` as a topic.
1. `terraform init && terraform plan`
1. Create PR, review plan, merge, approve plan.

### Deleting a repo

1. In `repositories.tf`, find the resource representing the repo to be deleted.
1. Remove the entire block.
1. `terraform init && terraform plan`
1. Create PR, review plan, merge, approve plan.

