# Nuon Extension: install-config-history

Browse the app branch configuration history associated with a Nuon install, diff
any two app configs, and search component config/build histories.

## Run locally

The Nuon CLI supplies API credentials and the currently selected install to an
extension. From this directory:

```bash
uv run nuon-ext-install-config-history --help
```

To exercise the normal extension environment without installing:

```bash
NUON_API_TOKEN=... \
NUON_ORG_ID=org... \
NUON_INSTALL_ID=ins... \
uv run nuon-ext-install-config-history
```

The repository experiment directory intentionally does not use the
`nuon-ext-<name>` basename required by `nuon ext install <local-directory>`.
For a local CLI install, create an alias with the expected basename:

```bash
ln -s "$PWD/exp/install-config-history" /tmp/nuon-ext-install-config-history
nuon ext install /tmp/nuon-ext-install-config-history
nuon install-config-history
```

Remove it with:

```bash
nuon ext remove install-config-history
rm /tmp/nuon-ext-install-config-history
```

## Interactive browser

```bash
nuon install-config-history
```

The Config history tab lists the selected install's branch configs, newest
first. When a config was produced by an app branch run, its row includes a
human-readable run summary (PR, tag, commit subject, or workflow type) next
to the config ID. Select a config to compare it with the preceding revision.
To compare arbitrary configs, highlight the baseline and press `b`, then
highlight the target and press `d`. Loading indicators appear while history,
diffs, and component details are fetched.

The Components tab searches by component name, type, or ID. Select a component
to load its configuration changes and builds.

Keys:

| Key | Action |
| --- | --- |
| `b` | Set the highlighted config as the diff baseline |
| `d` | Diff the highlighted config against the baseline |
| `/` | Focus component search |
| `r` | Refresh |
| `q` | Quit |

## Scriptable commands

```bash
nuon install-config-history history
nuon install-config-history history --output json

nuon install-config-history diff <old-config-id> <new-config-id>
nuon install-config-history diff <old-config-id> <new-config-id> --output json

nuon install-config-history component api
nuon install-config-history component <component-id> --output json
```

Global options can override CLI-provided context:

```bash
NUON_API_TOKEN=<api-token> uv run nuon-ext-install-config-history \
  --install-id <install-id> \
  --org-id <org-id> \
  history
```

All API calls are read-only.
