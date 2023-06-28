# Working in mono

## Main is golden

We follow the rule that `main` is golden. All changes happen in branches off of main.

We enforce a set of required CI checks which must pass before landing code. These CI checks are defined using IAC and can be found in `infra/github/repositories.tf`.

When code is merged to `main`, it is automatically deployed to `stage.

## Promotions

We work on a `promotion` basis - and have a manual workflow for deploying to `production`.

To trigger a promotion to prod

## Branch names + PR conventions

We require all branch names have an allowed prefix. We define the prefixes [here](https://github.com/powertoolsdev/mono/blob/main/.github/workflows/branch.yml#L19). Your initials should be your prefix for changes.

We currently do not enforce approvals, but have a protected branch. All PRs must have the appropriate PR title.
