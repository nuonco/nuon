* Modified from https://github.com/tibdex/github-app-token

# GitHub App Token

This [JavaScript GitHub Action](https://help.github.com/en/actions/building-actions/about-actions#javascript-actions)
can be used to impersonate a GitHub App when `secrets.GITHUB_TOKEN`'s
limitations are too restrictive and a personal access token is not suitable.

[`secrets.GITHUB_TOKEN`](https://help.github.com/en/actions/configuring-and-managing-workflows/authenticating-with-the-github_token)
has limitations such as [not being able to triggering a new workflow from another workflow](https://github.community/t5/GitHub-Actions/Triggering-a-new-workflow-from-another-workflow/td-p/31676).

Moreover, in an organization / enterprise, having workflows tied to an
individual's PAT is not a great idea for security or business continuity.

In many cases, using the default token in a workflow will be perfectly
sufficient. This action is really more appropriate for workflows in a repo that
need access to a different repo. Namely, our `go` services use this to consume
libraries in other repos.

# Example Workflow

```yml
jobs:
  job:
    runs-on: ubuntu-latest
    steps:
      - name: Generate token
        id: generate_token
        uses: powertoolsdev/mono/actions/app-token@main
        with:
          app_id: ${{ secrets.APP_ID }}
          installation_id: ${{ secrets.CROSS_REPO_INSTALLATION_ID }}
          private_key: ${{ secrets.PRIVATE_KEY }}
          repositories: go-common,another-nuon-repo

      - name: Checkout repo
        uses: actions/checkout@v3
        with:
          token: ${{ steps.generate_token.outputs.token }}
```
