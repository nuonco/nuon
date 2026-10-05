export default {
  title: 'Views / Apps / Installs',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/installs')} />
)

export const Results = () => page('installs')
export const Empty = () => page('installs-empty')
export const Loading = () => page('installs-loading')
