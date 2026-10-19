export default {
  title: 'Views / Installs / Deployment details',
  meta: { fullBleed: true, installViews: true },
}

import { InstallView } from './InstallView'
import { viewPath } from './install-fixture'
import { deploymentDetailFixture } from './page-fixtures'

const page = (
  state: Parameters<typeof deploymentDetailFixture>[0],
  tab = 'template-updates'
) => (
  <InstallView
    fixture={deploymentDetailFixture(state)}
    path={viewPath(`/deployments/wf-deploy-1/${tab}`)}
  />
)

export const InProgress = () => page('in-progress')
export const AwaitingApproval = () => page('awaiting')
export const Succeeded = () => page('succeeded')
export const Failed = () => page('failed')
export const InputUpdate = () => page('input-update', 'changes')
