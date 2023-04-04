# setup-helm

This action is a convenience wrapper for setting up helm for Nuon.

It installs the helm tool, installs a number of plugins, updates dependencies
and adds any additional repos.

```yml
    steps:
      - name: Setup helm
        uses: powertoolsdev/mono/actions/setup-helm@main
```

## Configuration

- `github_token`: Generally, does not need to be set
- `helm_dir`: The directory containing the chart. If the chart has
dependencies requiring auth, that will need to be setup first. (default: `k8s`)
- `repos`: Comma delimted list of name=url key values to add as repos.
(default: "")

## Plugins

- `local-chart-version`: Helper for interacting with chart versions.
- `s3`: Plugin for interacting with repos backed by S3
