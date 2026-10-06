export default {
  title: 'Views / Apps / Actions',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/actions')} />
)

export const Results = () => page('actions')
export const Empty = () => page('actions-empty')
export const Loading = () => page('actions-loading')
