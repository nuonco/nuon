export default {
  title: 'Views / Installs / Health',
  meta: { fullBleed: true, installViews: true },
}

import { InstallHealthPreview } from '@/components/installs/InstallHealth/InstallHealthPreview'
import { InstallView } from './InstallView'
import { NewInstallHealth } from './NewInstallHealth'
import { viewPath } from './install-fixture'
import { healthFixture } from './page-fixtures'

const page = (
  state: Parameters<typeof healthFixture>[0] | 'stale' | 'stale-checks'
) => (
  <InstallView
    fixture={healthFixture(
      state === 'stale' || state === 'stale-checks' ? 'healthy' : state
    )}
    path={viewPath('/health')}
    routeElements={{
      health: (
        <NewInstallHealth>
          <InstallHealthPreview state={state} />
        </NewInstallHealth>
      ),
    }}
  />
)

export const ResourceCanvas = () => page('degraded')
export const Healthy = () => page('healthy')
export const Degraded = () => page('degraded')
export const Unhealthy = () => page('unhealthy')
export const AccessError = () => page('access-error')
export const NoObservations = () => page('no-observations')
export const EmptyResources = () => page('empty')
export const Loading = () => page('loading')
export const StaleSnapshot = () => page('stale')
export const StaleChecks = () => page('stale-checks')
export const CurrentAPI = () => (
  <InstallView fixture={healthFixture('degraded')} path={viewPath('/health')} />
)
