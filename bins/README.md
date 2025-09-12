# bins

This is where we keep CLI applications.

## Waypoint Plugins

Even though these are used as part of our build and deploy pipelines, [Waypoint plugins are binaries](https://developer.hashicorp.com/waypoint/docs/extending-waypoint/plugin-interfaces), and are executed by ODRs as such. So we keep them here instead of in [services](../services).

### Adding a new plugin

To add a new Waypoint plugin, follow these steps:

1. Copy [waypoint-plugin-exp](waypoint-plugin-exp)
1. Update the BIN arg in the Earthfile
1. Add the infra for deploying the plugin by updating these files. For each file, you can copy the TF from another plugin and update the name.
   1. [infra/artifacts/ecr.tf](../infra/artifacts/ecr.tf)
   1. [infra/artifacts/github_actions.tf](../infra/artifacts/github_actions.tf)
   1. [infra/artifacts/outputs.tf](../infra/artifacts/outputs.tf)
1. Update [pkg/plugins](../pkg/plugins).
