export default {
  title: 'Views / Apps / Sandboxes',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/sandbox')} />
)

export const Configured = () => page('sandbox')
export const Unconfigured = () => page('sandbox-unconfigured')
export const Loading = () => page('sandbox-loading')
