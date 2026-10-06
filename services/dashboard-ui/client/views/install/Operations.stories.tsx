export default {
  title: 'Views / Installs / Operations',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { operationsFixture } from './install-fixtures'

export const Empty = () => (
  <InstallView fixture={operationsFixture('empty')} path={viewPath('/operations')} />
)

export const Active = () => (
  <InstallView fixture={operationsFixture('active')} path={viewPath('/operations')} />
)

export const RunnerOffline = () => (
  <InstallView
    fixture={operationsFixture('offline')}
    path={viewPath('/operations')}
  />
)
