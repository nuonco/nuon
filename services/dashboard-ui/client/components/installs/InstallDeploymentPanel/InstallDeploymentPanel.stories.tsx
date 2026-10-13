export default {
  title: 'Features / Installs / Install deployment panel',
  fullBleed: true,
}

import { DeploymentDetailLoading } from '@/components/installs/DeploymentsList/DeploymentDetailLoading'
import { Panel } from '@/components/surfaces/Panel'

export const Loading = () => (
  <Panel
    heading="payments"
    isVisible
    size="3/4"
    aria-label="Deployment details"
  >
    <DeploymentDetailLoading />
  </Panel>
)
