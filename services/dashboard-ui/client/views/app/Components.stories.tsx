export default {
  title: 'Views / Apps / Components',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/components')} />
)

export const Results = () => page('components')
export const Empty = () => page('components-empty')
export const Loading = () => page('components-loading')
