export default {
  title: 'Views / Installs / Deployments',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { deploymentsFixture } from './page-fixtures'

export const Loading = () => (
  <InstallView
    fixture={deploymentsFixture('loading')}
    path={viewPath('/deployments')}
  />
)

export const Empty = () => (
  <InstallView
    fixture={deploymentsFixture('empty')}
    path={viewPath('/deployments')}
  />
)

export const Results = () => (
  <InstallView
    fixture={deploymentsFixture('results')}
    path={viewPath('/deployments')}
  />
)

export const AwaitingApproval = () => (
  <InstallView
    fixture={deploymentsFixture('results')}
    path={viewPath('/deployments?active=awaiting-approval')}
  />
)

export const FailedPendingRetry = () => (
  <InstallView
    fixture={deploymentsFixture('results')}
    path={viewPath('/deployments?active=failed')}
  />
)

export const HistoryFiltered = () => (
  <InstallView
    fixture={deploymentsFixture('results')}
    path={viewPath('/deployments?status=failed,cancelled')}
  />
)
