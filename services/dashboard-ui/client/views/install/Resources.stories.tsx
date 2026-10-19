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
  managedByConfig = false
) => (
  <InstallView
    fixture={resourcesFixture('active', alignment, managedByConfig)}
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
export const ComponentsTornDown = () =>
  aligned('/resources/components', 'torn-down')
export const ImagesPendingSync = () => aligned('/resources/images', 'drifted')
export const ImagesSyncing = () => aligned('/resources/images', 'ongoing')
export const ImagesSyncCancelled = () =>
  aligned('/resources/images', 'image-sync-cancelled')
export const ImagesSynced = () => aligned('/resources/images', 'in-sync')
export const SandboxDrifted = () => aligned('/resources/sandbox', 'drifted')
export const SandboxOngoing = () => aligned('/resources/sandbox', 'ongoing')
export const StackInSync = () => aligned('/resources', 'drifted')
export const AlignmentLoadFailed = () =>
  aligned('/resources/components', 'error')
