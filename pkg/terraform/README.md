# Terraform

This package exposes the ability to work with `terraform` locally.

## Concepts

* `workspace` - stolen from `terraform cloud`, a workspace represents the environment for running terraform, and various
  commands.
* `variables` - used to set input variables into a terraform workspace
* `backend` - used to configure a backend
* `run` - a run is a single plan, delete, or apply
* `archive` -  represents a set of source or other files
