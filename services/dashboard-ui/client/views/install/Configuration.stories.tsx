export default {
  title: 'Views / Installs / Configuration',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { configurationFixture } from './page-fixtures'

const page = (state: Parameters<typeof configurationFixture>[0]) => (
  <InstallView fixture={configurationFixture(state)} path={viewPath('/configuration')} />
)

export const UnsetBranch = () => page('unset')
export const Manual = () => page('manual')
export const ConfigManaged = () => page('managed')
