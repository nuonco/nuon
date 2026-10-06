export default {
  title: 'Views / Apps / Settings',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/settings')} />
)

export const WithPlan = () => page('settings')
export const NoPlan = () => page('settings-no-plan')
