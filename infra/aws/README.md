# infra-aws

## Concepts

Looking at `main.tf` or any of the stack configs in `./lib/`, one may presume
that the CDK is creating resources as each class is instantiated. This is not
the case. Running the program will synthesize the resources and generate a
JSON file for each stack that can be planned/applied.

It is possible to use advanced logic during synthesis; however, one must always
keep in mind when the logic runs - either during synthesis or during apply.

## Setup

- Install dependencies: `npm ci`
- Synthesize `npm run synth`

## Changing infrastructure

Terraform cloud is used to run `terraform plan` and `terraform apply` on the synthesized terraform that `cdktf` creates. You should _not_ run `cdktf` commands locally, as they can easily delete resources.

- `npm run synth`

# Configuration

The root account is enrolled in organizations and is the "root" account.
Resources generally should not be created here. They should be created in the
appropriate sub-account.


# Organizational Units(OU) and Accounts

OUs are a mechanism for organizing subaccounts. Policies should generally be
attached here. Accounts are where resources are actually deployed.


## `meta`

This OU is for accounts that are responsible for managing our AWS accounts.
Has a reasonable set of policies attached to try to prevent anything untoward
from happening.

#### `audit`

Currently used for shared storage of org-wide CloudTrail logs.

#### `external`

Intended for anything that customers may rely on - e.g. roles. This account is
intentionally segregated from any of our workloads.

#### `logs`

Not currently used but intended for storing permenent log data.

#### `security`

Not currently used but intended for use by a future security team.

## `nuon-testing`

This OU is intended for testing our product. Policies aren't currently applied
to accounts under this OU so you may create an account and creds for testing.

## `workloads`

This OU is intended for our workloads e.g. running the API and app.
Has a reasonable set of policies attached to try to prevent anything untoward
from happening.

The desire is that accounts should be as small as possible without requiring an
onerous amount of cross-account permissions or VPC peering.

#### `sandbox`

For messing around in, testing, etc. In the future, we'll run a periodic job to
cleanup any stale resources but, for now, clean up after yourself.

#### `stage`

The primary staging account.

#### `infra-shared-stage`

Not currently used but intended for shared infrastructure that would be used
by multiple "staging" accounts.


#### `prod`

The primary production account.

#### `infra-shared-prod`

Intended for shared infrastructure used by multiple "production" accounts.
Currently, our helm chart and ECR repositories live there.

#### `orgs-stage`

Intended to run and store all org infrastructure, such as waypoint servers and runners + build artifacts.

#### `orgs-prod`

Intended to run and store all org infrastructure, such as waypoint servers and runners + build artifacts.

# SSO

We're currently using AWS SSO to provision access. At a high level, we're
syncing users and groups from GSuite to AWS, granting permission sets
to those groups and attaching those groups to all of our subaccounts.

### Enablement

SSO has to be enabled from the console. Fortunately, this is a one-time thing
only in the root / management account.

### Permission sets and attachments

These are created here in Terraform.

### User / Group sync

This requires set up in Google Cloud console, Google Workspaces Admin console,
as well as the AWS SSO console. Finally, there's a Lambda that queries Google
Workspaces and pushes users to AWS.

#### Google Cloud

A new project was created to enable the Admin SDK and create a service account.

https://console.cloud.google.com/home/dashboard?project=aws-sso-idp-sync

#### Google Workspaces

After creating the service account, we have to provide domain-wide permissions
to the service account. This gives the service account the ability to read user
and group information without each user having to give permission for those
OAuth scopes.

https://admin.google.com/ac/owl/domainwidedelegation?hl=en

#### SSO SCIM

Again, enabling automatic provisioning is only available via the UI. We create
a SCIM endpoint and access token in the SSO external identity provider
settings.

#### Lambda

This is deployed as a Serverless Application outside of Terraform as it's a
one-off.

https://us-west-2.console.aws.amazon.com/lambda/home?region=us-west-2#/applications/serverlessrepo-idp-scim-sync

