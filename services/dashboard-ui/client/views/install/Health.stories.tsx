export default {
  title: 'Views / Installs / Health',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { healthFixture } from './page-fixtures'

const page = (state: Parameters<typeof healthFixture>[0]) => (
  <InstallView fixture={healthFixture(state)} path={viewPath('/health')} />
)

export const Healthy = () => page('healthy')
export const Degraded = () => page('degraded')
export const Unhealthy = () => page('unhealthy')
export const AccessError = () => page('access-error')
export const NoObservations = () => page('no-observations')
export const EmptyResources = () => page('empty')
export const Loading = () => page('loading')
