export default {
  title: 'Playground / Installs / Deployments',
}

import { DeploymentsPreview } from '@/components/installs/DeploymentDetail/DeploymentDetail.preview'
import { LayoutControls as LayoutControlsPreview } from './DeploymentRowLayout.preview'

export const InteractiveSandbox = () => (
  <DeploymentsPreview scenario="interactive-sandbox" />
)

export const LayoutControls = () => <LayoutControlsPreview />
