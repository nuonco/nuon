export default {
  title: 'Views / Apps / Pull requests',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

export const Open = () => (
  <InstallView
    fixture={appFixture('pull-request')}
    path={appPath('/branches/br-1/pull-requests')}
  />
)

export const Loading = () => (
  <InstallView
    fixture={appFixture('pull-request-loading')}
    path={appPath('/branches/br-1/pull-requests')}
  />
)
