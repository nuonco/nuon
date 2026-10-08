export default {
  title: 'Playground / Installs / Deployments',
}

import { DeploymentsPreview } from '@/components/installs/DeploymentDetail/DeploymentDetail.preview'
import { LayoutControls as LayoutControlsPreview } from './DeploymentRowLayout.preview'
import { DeploymentPoliciesPreview } from './DeploymentPolicies.preview'

export const InteractiveSandbox = () => (
  <DeploymentsPreview scenario="interactive-sandbox" />
)

export const LayoutControls = () => <LayoutControlsPreview />

export const PolicyVisibility = () => <DeploymentPoliciesPreview />
