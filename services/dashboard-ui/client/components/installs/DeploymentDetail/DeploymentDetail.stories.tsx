export default {
  title: 'Features / Installs / Deployment details',
}

import { DeploymentsPreview } from './DeploymentDetail.preview'

export const RunningComponents = () => (
  <DeploymentsPreview scenario="running-components" />
)

export const ComponentsNotStarted = () => (
  <DeploymentsPreview scenario="components-not-started" />
)

export const PendingApproval = () => (
  <DeploymentsPreview scenario="pending-approval" />
)

export const PartialFailure = () => (
  <DeploymentsPreview scenario="partial-failure" />
)

export const StackRecovery = () => (
  <DeploymentsPreview scenario="stack-recovery" />
)

export const Completed = () => <DeploymentsPreview scenario="completed" />
