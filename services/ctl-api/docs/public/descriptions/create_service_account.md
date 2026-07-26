Create a service account for the current org. Service accounts can be used to
generate API tokens for automation and CI/CD workflows.

Defaults to the `org_admin` role if `role` is not specified. Allowed roles
are `org_admin`, `installer`, and `runner`.
