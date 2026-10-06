export default {
  title: 'Views / Installs / Resources',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { resourcesFixture } from './page-fixtures'

const page = (state: Parameters<typeof resourcesFixture>[0]) => (
  <InstallView fixture={resourcesFixture(state)} path={viewPath('/resources')} />
)

export const Empty = () => page('empty')
export const Provisioning = () => page('provisioning')
export const Active = () => page('active')
export const Degraded = () => page('degraded')
