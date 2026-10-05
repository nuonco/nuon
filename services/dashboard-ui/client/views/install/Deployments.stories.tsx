export default {
  title: 'Views / Installs / Deployments',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { deploymentsFixture } from './page-fixtures'

export const Loading = () => (
  <InstallView fixture={deploymentsFixture('loading')} path={viewPath('/deployments')} />
)

export const Empty = () => (
  <InstallView fixture={deploymentsFixture('empty')} path={viewPath('/deployments')} />
)

export const Results = () => (
  <InstallView fixture={deploymentsFixture('results')} path={viewPath('/deployments')} />
)

export const Filtered = () => (
  <InstallView
    fixture={deploymentsFixture('results')}
    path={viewPath('/deployments?search=worker')}
  />
)
