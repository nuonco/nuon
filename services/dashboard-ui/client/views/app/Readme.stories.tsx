export default {
  title: 'Views / Apps / README',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from '@/views/install/InstallView'
import { appFixture, appPath } from './app-fixtures'

const page = (state: string) => (
  <InstallView fixture={appFixture(state)} path={appPath('/branches/br-1/readme')} />
)

export const Rendered = () => page('readme')
export const Empty = () => page('readme-empty')
export const Loading = () => page('readme-loading')
