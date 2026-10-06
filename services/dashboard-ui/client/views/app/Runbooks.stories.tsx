export default {
  title: 'Views / Apps / Runbooks',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/runbooks')} />
)

export const Results = () => page('runbooks')
export const Empty = () => page('runbooks-empty')
export const Loading = () => page('runbooks-loading')
