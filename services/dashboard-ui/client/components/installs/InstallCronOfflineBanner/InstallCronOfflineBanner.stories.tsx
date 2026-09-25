export default {
  title: 'Installs/InstallCronOfflineBanner',
}

import { InstallCronOfflineBanner } from './InstallCronOfflineBanner'

export const ActionCrons = () => (
  <InstallCronOfflineBanner
    runnerStatus="offline"
    kind="action"
    hasCronSchedule
    orgId="org_example"
    installId="inst_example"
  />
)

export const SandboxDrift = () => (
  <InstallCronOfflineBanner
    runnerStatus="offline"
    kind="sandbox_drift"
    hasCronSchedule
    orgId="org_example"
    installId="inst_example"
  />
)

export const ComponentDrift = () => (
  <InstallCronOfflineBanner
    runnerStatus="offline"
    kind="component_drift"
    hasCronSchedule
    orgId="org_example"
    installId="inst_example"
  />
)

export const HiddenWhenRunnerActive = () => (
  <InstallCronOfflineBanner
    runnerStatus="active"
    kind="action"
    hasCronSchedule
    orgId="org_example"
    installId="inst_example"
  />
)

export const HiddenWithoutCronSchedule = () => (
  <InstallCronOfflineBanner
    runnerStatus="offline"
    kind="sandbox_drift"
    hasCronSchedule={false}
    orgId="org_example"
    installId="inst_example"
  />
)
