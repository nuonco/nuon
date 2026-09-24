export default {
  title: 'Installs/InstallCronOfflineBanner',
}

import { InstallCronOfflineBanner } from './InstallCronOfflineBanner'

export const ActionCrons = () => (
  <InstallCronOfflineBanner
    runnerStatus="offline"
    kind="action"
    hasCronSchedule
  />
)

export const SandboxDrift = () => (
  <InstallCronOfflineBanner
    runnerStatus="offline"
    kind="sandbox_drift"
    hasCronSchedule
  />
)

export const ComponentDrift = () => (
  <InstallCronOfflineBanner
    runnerStatus="offline"
    kind="component_drift"
    hasCronSchedule
  />
)

export const HiddenWhenRunnerActive = () => (
  <InstallCronOfflineBanner
    runnerStatus="active"
    kind="action"
    hasCronSchedule
  />
)

export const HiddenWithoutCronSchedule = () => (
  <InstallCronOfflineBanner
    runnerStatus="offline"
    kind="sandbox_drift"
    hasCronSchedule={false}
  />
)
