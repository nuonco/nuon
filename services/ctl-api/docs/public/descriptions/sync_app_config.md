Sync an app config that was created with an intermediate config.

The config is applied asynchronously: this returns `202` immediately and the
config moves through `syncing` to `active` or `error`. Poll
`GET /v1/apps/{app_id}/configs/{config_id}` for the outcome — `status`,
`status_description`, and the resolved `component_ids` / `action_ids` /
`runbook_ids`. Scheduled component builds and resources orphaned by this sync are
reported under `state.result`.

Component builds are scheduled as part of the sync. A component whose config is
unchanged since the previous sync, and whose last build did not fail, keeps its
existing config connection and is not rebuilt.
