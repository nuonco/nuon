Returns CLI configuration and settings.

`min_cli_version` is deprecated. It remains the warning floor for older CLIs. Prefer `min_cli_version_policy.warn`. `min_cli_version_policy.error` is the oldest CLI this control plane will run. An empty warn means the CLI does not warn.
