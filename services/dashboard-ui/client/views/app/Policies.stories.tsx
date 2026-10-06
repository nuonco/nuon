export default {
  title: 'Views / Apps / Policies',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/policies')} />
)

export const Results = () => page('policies')
export const Empty = () => page('policies-empty')
export const Loading = () => page('policies-loading')
