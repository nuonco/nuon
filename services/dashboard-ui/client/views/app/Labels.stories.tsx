export default {
  title: 'Views / Apps / Labels',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/labels')} />
)

export const Results = () => page('labels')
export const Empty = () => page('labels-empty')
export const Loading = () => page('labels-loading')
