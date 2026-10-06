export default {
  title: 'Views / Apps / Branches',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath()} />
)

export const Several = () => page('picker')
export const Empty = () => page('picker-empty')
export const Loading = () => page('picker-loading')
