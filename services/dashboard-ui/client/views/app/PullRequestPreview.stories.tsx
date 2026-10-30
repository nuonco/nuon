export default {
  title: 'Views / Apps / Pull request preview',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

export const Default = () => (
  <InstallView
    fixture={appFixture('pull-request')}
    path={appPath('/branches/br-1/pull-requests/482')}
  />
)

export const Loading = () => (
  <InstallView
    fixture={appFixture('pull-request-loading')}
    path={appPath('/branches/br-1/pull-requests/482')}
  />
)

export const NotFound = () => (
  <InstallView
    fixture={appFixture('pull-request')}
    path={appPath('/branches/br-1/pull-requests/999')}
  />
)
