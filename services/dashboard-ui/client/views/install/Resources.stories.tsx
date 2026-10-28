export default {
  title: 'Views / Installs / Resources',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { resourcesFixture } from './page-fixtures'

const page = (state: Parameters<typeof resourcesFixture>[0]) => (
  <InstallView
    fixture={resourcesFixture(state)}
    path={viewPath('/resources')}
  />
)

export const Empty = () => page('empty')
export const Provisioning = () => page('provisioning')
export const Active = () => page('active')
export const Degraded = () => page('degraded')

const aligned = (
  suffix: string,
  alignment: Parameters<typeof resourcesFixture>[1],
  managedByConfig = false,
  resolutionScenario?: Parameters<typeof resourcesFixture>[3]
) => (
  <InstallView
    fixture={resourcesFixture(
      'active',
      alignment,
      managedByConfig,
      resolutionScenario
    )}
    path={viewPath(suffix)}
  />
)

export const ComponentsDrifted = () =>
  aligned('/resources/components', 'drifted')
export const ComponentsOngoing = () =>
  aligned('/resources/components', 'ongoing')
export const ComponentsDriftedConfigManaged = () =>
  aligned('/resources/components', 'drifted', true)
export const ComponentsInputUpdates = () =>
  aligned('/resources/components', 'input-updates')
export const ComponentsInputUpdatesConfigManaged = () =>
  aligned('/resources/components', 'input-updates', true)
export const ComponentsInSync = () =>
  aligned('/resources/components', 'in-sync')
export const ComponentsInputsUnverified = () =>
  aligned('/resources/components', 'inputs-unverified')
export const ComponentsTornDown = () =>
  aligned('/resources/components', 'torn-down')
export const ComponentsDeploymentHistory = () =>
  aligned('/resources/components', 'component-history')
export const ComponentsPaged = () => aligned('/resources/components', 'paged')
export const ImagesPendingSync = () => aligned('/resources/images', 'drifted')
export const ImagesSyncing = () => aligned('/resources/images', 'ongoing')
export const ImagesSyncCancelled = () =>
  aligned('/resources/images', 'image-sync-cancelled')
export const ImagesSynced = () => aligned('/resources/images', 'in-sync')
export const ImagesTornDown = () => aligned('/resources/images', 'torn-down')
export const ComponentsAlignmentUnavailable = () =>
  aligned('/resources/components', 'unavailable')
export const ImagesAlignmentUnavailable = () =>
  aligned('/resources/images', 'unavailable')
export const AlignmentLoadFailed = () =>
  aligned('/resources/components', 'error')

export const ResolveChangedPlan = () =>
  aligned('/resources/components', 'drifted', false, 'changed-plan')
export const ResolveAutoApproved = () =>
  aligned('/resources/components', 'drifted', false, 'auto-approved')
export const ResolvePolicyDenied = () =>
  aligned('/resources/components', 'drifted', false, 'policy-denied')
