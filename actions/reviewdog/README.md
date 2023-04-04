# reviewdog

This action runs a standard set of linters. Generally, this action should be
able to be used with minimal setup on every repository.

```yml
    steps:
      - name: Lint
        uses: powertoolsdev/mono/actions/reviewdog@main
```

## Linters

All of the linters are wrapped by reviewdog which handles interfacing with
Github so that we get PR comments and suggestions instead of pass/fail without
any additional feedback.

### actionlint

Lints github actions (e.g. `./.github/workflows/*.yml`). Runs a number of checks
to validate that the workflow is correct including yml formatting, shell script
checking, etc.

### yamllint

Runs yamllint. Each repo can currently specify a `./.yamllint.yaml` file to
configure the linter.

### shellcheck

Runs shellcheck to validate shell scripts. If there are no shell scripts, it
does not fail.

### terraform

Runs `tf lint`, `terraform validate` and `terraform fmt`. On PRs, `fmt` results
will be presented as suggested changes in PR comments that can be accepted
inline.

Terraform linting can be turned off for repos that don't contain terraform.

In order for `terraform validate` to run correctly, `terraform init` must be
called beforehand. For many repos, calling it without initing the backend is
appropriate - `terraform init -backend=false`.

### hadolint

Hadolint validates both the correctness and security of Dockerfiles. If a repo
includes one, it should enable this linter.

### node

Runs `npm run lint`. Each repo using TS/Node can configure linting and provide a "lint" script in the `package.json`.


## Configuration

- `github_token`: Generally does not need to be set as it will default to the
GHA token.
- `run_all`: Whether to run all linters or to stop on first failure. (default: `"true"`)
- `terraform`: Whether to run terraform checks. `terraform init` must be called
before invoking (default: `"true"`)
- `terraform_dir`: The directory containing terraform. (default: `infra`)
- `hadolint`: Whether to run hadolint. (default: `"false"`)


## TODO

The actions in the reviewdog org aren't particularly very good. It may behoove
us to create a container with all of the tools and use it instead.
