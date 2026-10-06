export default {
  title: 'Views / Installs / Overview',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { overviewFixture } from './page-fixtures'

export const NotReady = () => (
  <InstallView fixture={overviewFixture('empty')} path={viewPath('/')} />
)

export const WithWarnings = () => (
  <InstallView fixture={overviewFixture('warnings')} path={viewPath('/')} />
)

export const Rendered = () => (
  <InstallView fixture={overviewFixture('rendered')} path={viewPath('/')} />
)
