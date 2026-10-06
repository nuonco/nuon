export default {
  title: 'Views / Apps / Inputs',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/inputs')} />
)

export const Defined = () => page('inputs')
export const Empty = () => page('inputs-empty')
export const MissingConfig = () => page('inputs-missing')
export const Failed = () => page('inputs-failed')
export const Loading = () => page('inputs-loading')
