# Infrastructure

This code manages all of our infra.

## File Organization

An easy way to understand how it all fits together is to think of it in layers:

1. `/terraform`: manages terraform workspaces.
1. `/aws`: manages the AWS accounts and SSO.
1. `/eks`: manages EKS clusters.
1. `/orgs`: expose EKS fields and settings that are used by the workers.

The terraform in the rest of the directories runs on top of all that, to provision common services like Datadog, Temporal, and Waypoint. Finally, the terraform in each service directory runs on top of everything here, to provision app-specific resources.

## Workspaces

We use workspaces to organize our Terraform code. These are, themselves, configured using Terraform, and you can find that code in `/terraform`.

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


## Architecture

Our infrastructure is all in AWS, with the exception of our IdP, which is Auth0. Within AWS, we have a pretty standard EKS architecture, with a few wrinkles that mainly stem from isolating org infra away from our own service infra.

### AWS Accounts

We split up our infra across 3 AWS accounts, primarily to isolate org-specific infrastructure from our own, internal infra.

1. prod: Hosts Nuon services, like our API.
1. orgs-prod: Hosts org-specific services, like the Waypoint servers and runners created for each org.
1. infra-shared-prod: (I'm actually not sure what this is for yet.)

We also have a staging env with the same structure, but with `prod` swapped for `stage` in all the names.

TODO: Outline the other accounts, like canary.

### AWS Infra

The prod account contains a pretty standard EKS setup. (Insert some Miro diagrams here, like this one: https://miro.com/app/board/uXjVMT5BLxs=/?moveToWidget=3458764552294864172&cot=14)
