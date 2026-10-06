export default {
  title: 'Views / Apps / Previous runs',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/runs')} />
)

export const Results = () => page('runs')
export const Failed = () => page('runs-failed')
export const Empty = () => page('runs-empty')
export const Loading = () => page('runs-loading')
