import {
  ACTIVE_DEPLOYMENT_STATUSES,
  DEPLOYMENT_TABS,
  deploymentTabOrder,
  type TDeploymentTab,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import type { TNavLink, TWorkflow } from '@/types'

export type TDeploymentV2Tab = TDeploymentTab | 'live'

export const DEPLOYMENT_V2_TABS: Record<TDeploymentV2Tab, TNavLink> = {
  ...DEPLOYMENT_TABS,
  live: { path: '/live', text: 'Run view' },
}

export const isWorkflowRunning = (workflow?: TWorkflow) =>
  ACTIVE_DEPLOYMENT_STATUSES.has(workflow?.status?.status ?? '')

export const deploymentV2TabOrder = (
  workflow?: TWorkflow
): TDeploymentV2Tab[] => {
  const base = deploymentTabOrder(workflow?.type)
  if (isWorkflowRunning(workflow)) return ['live', ...base]
  return ['changes', ...base.filter((key) => key !== 'changes'), 'live']
}

export const deploymentV2Tabs = (workflow?: TWorkflow): TNavLink[] =>
  deploymentV2TabOrder(workflow).map((key) => DEPLOYMENT_V2_TABS[key])
