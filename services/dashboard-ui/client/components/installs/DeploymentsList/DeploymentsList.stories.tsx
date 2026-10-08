import { useState } from 'react'
import { PanelStory } from '@/components/__stories__/helpers'
import type {
  TInstallDeploymentStepSummary,
  TInstallDeploymentSummary,
} from '@/types'
import { DeploymentDetailPanel } from './DeploymentDetailPanel'
import {
  DeploymentsListPresenter,
  type IDeploymentFilter,
} from './DeploymentsListPresenter'

// ─── Fixtures ────────────────────────────────────────────────────────────────

const ORG_ID = 'org_example'
const APP_ID = 'app_example'
const INSTALL_ID = 'install_example'

const step = (
  id: string,
  name: string,
  status: string,
  group_idx: number,
  props: Partial<TInstallDeploymentStepSummary> = {}
): TInstallDeploymentStepSummary => ({
  id,
  name,
  status,
  group_idx,
  execution_type: 'system',
  ...props,
})

const DEPLOYMENT_PROVISION: TInstallDeploymentSummary = {
  id: 'dep_provision_1',
  type: 'provision',
  status: 'success',
  created_at: new Date(Date.now() - 7 * 24 * 3600_000).toISOString(),
  title: 'Initial provision',
  activity: 'All components provisioned for the first time.',
  finished: true,
  steps: [
    step('s1', 'await install stack', 'success', 1, {
      step_target_type: 'install_stack_versions',
    }),
    step('s2', 'provision sandbox apply plan', 'success', 2, {
      step_target_type: 'install_sandbox_runs',
    }),
    step('s3', 'apply api', 'success', 3, { component_name: 'api' }),
    step('s4', 'apply worker', 'success', 4, { component_name: 'worker' }),
  ],
}

const DEPLOYMENT_COMPONENT_DEPLOY: TInstallDeploymentSummary = {
  id: 'dep_component_2',
  type: 'component_deploy',
  status: 'in-progress',
  created_at: new Date(Date.now() - 2 * 3600_000).toISOString(),
  title: 'Deploy api image',
  activity: 'Applying api',
  steps: [
    step('s1', 'plan api', 'success', 1, { component_name: 'api' }),
    step('s2', 'apply api', 'in-progress', 2, { component_name: 'api' }),
    step('s3', 'verify install health', 'pending', 3),
  ],
}

const DEPLOYMENT_CONFIG_UPDATE: TInstallDeploymentSummary = {
  id: 'dep_config_3',
  type: 'install_config_update',
  status: 'error',
  created_at: new Date(Date.now() - 48 * 3600_000).toISOString(),
  title: 'Config update',
  activity: 'Applying worker failed',
  finished: true,
  steps: [
    step('s1', 'apply api', 'success', 1, { component_name: 'api' }),
    step('s2', 'apply worker', 'error', 2, { component_name: 'worker' }),
  ],
}

const DEFAULT_FILTER: IDeploymentFilter = {
  search: '',
  status: new Set(),
  type: new Set(),
}

const MOCK_DEPLOYMENTS = [DEPLOYMENT_PROVISION, DEPLOYMENT_CONFIG_UPDATE]

const ACTIVE_DEPLOYMENTS: TInstallDeploymentSummary[] = [
  {
    ...DEPLOYMENT_COMPONENT_DEPLOY,
    id: 'dep_approval',
    title: 'Deploy api + worker',
    status: 'approval-awaiting',
    steps: [
      step('a1', 'await install stack', 'success', 1),
      step('a2', 'plan api', 'approval-awaiting', 2, {
        execution_type: 'approval',
        component_name: 'api',
      }),
      step('a3', 'apply api', 'pending', 3, { component_name: 'api' }),
    ],
  },
  {
    ...DEPLOYMENT_COMPONENT_DEPLOY,
    id: 'dep_retry',
    title: 'Deploy components',
    status: 'failed-pending-retry',
    activity: 'Apply worker failed — retry or skip',
    steps: [
      step('r1', 'apply api', 'success', 1, { component_name: 'api' }),
      step('r2', 'apply worker', 'error', 2, { component_name: 'worker' }),
      step('r3', 'verify install health', 'pending', 3),
    ],
  },
  {
    ...DEPLOYMENT_PROVISION,
    id: 'dep_provision_active',
    title: 'Provision',
    status: 'in-progress',
    finished: false,
    steps: [
      step('p1', 'await install stack', 'success', 1),
      step('p2', 'provision sandbox apply plan', 'in-progress', 2),
      step('p3', 'apply api', 'pending', 3, { component_name: 'api' }),
    ],
  },
  {
    ...DEPLOYMENT_COMPONENT_DEPLOY,
    id: 'dep_secrets',
    type: 'install_config_update',
    title: 'Sync secrets',
    steps: [step('secrets', 'sync secrets', 'in-progress', 1)],
  },
  ...Array.from({ length: 6 }, (_, index) => ({
    ...DEPLOYMENT_COMPONENT_DEPLOY,
    id: `dep_queued_${index}`,
    title: `Deploy queued component ${index + 1}`,
    status: 'pending' as const,
    steps: [],
  })),
]

// ─── Stories ─────────────────────────────────────────────────────────────────

export default {
  title: 'Features / Installs / Deployments list',
}

export const Default = () => {
  const [limit, setLimit] = useState(4)
  return (
    <div className="max-w-7xl mx-auto p-6">
      <DeploymentsListPresenter
        deployments={MOCK_DEPLOYMENTS}
        activeDeployments={ACTIVE_DEPLOYMENTS.slice(0, limit)}
        activeTotal={ACTIVE_DEPLOYMENTS.length}
        hasMoreActive={limit < ACTIVE_DEPLOYMENTS.length}
        onLoadMoreActive={() => setLimit((limit) => limit + 4)}
        isLoading={false}
        error={null}
        pagination={{ hasNext: false, offset: 0, limit: 20 }}
        orgId={ORG_ID}
        appId={APP_ID}
        installId={INSTALL_ID}
        search=""
        filter={DEFAULT_FILTER}
        onSearchChange={() => {}}
        onStatusChange={() => {}}
        onTypeChange={() => {}}
        onResourceChange={() => {}}
        onDateChange={() => {}}
        onClearFilters={() => {}}
      />
    </div>
  )
}

export const Loading = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={[]}
      isLoading
      activeLoading
      error={null}
      pagination={{ hasNext: false, offset: 0, limit: 5 }}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      search=""
      filter={DEFAULT_FILTER}
      onSearchChange={() => {}}
      onStatusChange={() => {}}
      onTypeChange={() => {}}
      onResourceChange={() => {}}
      onDateChange={() => {}}
      onClearFilters={() => {}}
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={[]}
      isLoading={false}
      error={null}
      pagination={{ hasNext: false, offset: 0, limit: 20 }}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      search=""
      filter={DEFAULT_FILTER}
      onSearchChange={() => {}}
      onStatusChange={() => {}}
      onTypeChange={() => {}}
      onResourceChange={() => {}}
      onDateChange={() => {}}
      onClearFilters={() => {}}
    />
  </div>
)

export const EmptyFiltered = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={[]}
      isLoading={false}
      error={null}
      pagination={{ hasNext: false, offset: 0, limit: 20 }}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      search=""
      filter={{ ...DEFAULT_FILTER, status: new Set(['failed']) }}
      onSearchChange={() => {}}
      onStatusChange={() => {}}
      onTypeChange={() => {}}
      onResourceChange={() => {}}
      onDateChange={() => {}}
      onClearFilters={() => {}}
    />
  </div>
)

export const WithPagination = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={MOCK_DEPLOYMENTS}
      isLoading={false}
      error={null}
      pagination={{ hasNext: true, offset: 20, limit: 20 }}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      search=""
      filter={DEFAULT_FILTER}
      onSearchChange={() => {}}
      onStatusChange={() => {}}
      onTypeChange={() => {}}
      onResourceChange={() => {}}
      onDateChange={() => {}}
      onClearFilters={() => {}}
    />
  </div>
)

export const DetailPanelStory = () => (
  <PanelStory label="Open deployment details">
    <DeploymentDetailPanel
      deployment={DEPLOYMENT_COMPONENT_DEPLOY}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
    />
  </PanelStory>
)

export const DetailPanelProvision = () => (
  <PanelStory label="Open provision details">
    <DeploymentDetailPanel
      deployment={DEPLOYMENT_PROVISION}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
    />
  </PanelStory>
)
