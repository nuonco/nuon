# infra

This directory contains infrastructure as code that powers `nuon`.

## earthly

We use `earthly` during CI to run various steps and validations against code changes here. To run this locally, do the following:

Export a terraform cloud team token with access to the `launchpaddev` workspace:

```bash
export EARTHLY_SECRETS="TERRAFORM_CLOUD_TOKEN=<your-token>"
```

Run earthly by passing in a module, as well as a workspace to be used for the backend during initialization:

```bash
$ earthly +lint --MODULE=datadog --TERRAFORM_WORKSPACE=infra-datadog-orgs-stage
```

## Key workspaces

### aws

Manages our AWS accounts, root account and SSO. This is built using `cdktf`.

### eks

Manages all compute infrastructure in accounts -- including prod, stage, orgs-stage, and orgs-prod. This module provisions VPCs, k8s, twingate and DNS infra.

### terraform

Manages our terraform cloud workspaces.

### orgs

Manages shared resources for orgs, and other things for interacting with the `orgs` accounts such as our support role. Services that provision infrastructure in the orgs account should use this instead of relying on `infra-eks-orgs-stage` directly.

### datadog

Manages the datadog agent via helm in our 4 primary accounts.

### dns

Manages root DNS for `nuon.co` domain.

### artifacts

Manages resources in the `public` account for pushing artifacts.

### github

Manages our github enterprise, repos, teams and configuration.

### temporal

Manages temporal via helm in our stage and prod accounts.
