export default {
  title: 'Views / Apps / Roles',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/roles')} />
)

export const Results = () => page('roles')
export const Empty = () => page('roles-empty')
export const Loading = () => page('roles-loading')
