import { PanelStory } from '@/components/__stories__/helpers'
import type { TInstallDeploymentRecord } from '@/types'
import { DeploymentDetailPanel } from './DeploymentDetailPanel'
import {
  DEFAULT_DEPLOYMENTS_FILTER,
  DeploymentsListPresenter,
} from './DeploymentsListPresenter'

// ─── Fixtures ────────────────────────────────────────────────────────────────

const ORG_ID = 'org_example'
const APP_ID = 'app_example'
const INSTALL_ID = 'install_example'

const DEPLOYMENT_PROVISION: TInstallDeploymentRecord = {
  id: 'dep_provision_1',
  type: 'provision',
  status: 'success',
  created_at: new Date(Date.now() - 7 * 24 * 3600_000).toISOString(),
  title: 'Initial provision',
  summary: 'All components provisioned for the first time.',
  app_branch: { id: 'branch_1', name: 'main', sha: 'a1b2c3d4e5f6' },
  workflow: { id: 'wf_1', name: 'Provision workflow', type: 'provision' },
  affected_resources: {
    stack: true,
    sandbox: true,
    components: ['api', 'worker'],
    images: [],
  },
  change_groups: [],
}

const DEPLOYMENT_IMAGE_UPDATE: TInstallDeploymentRecord = {
  id: 'dep_image_2',
  type: 'image_update',
  status: 'in-progress',
  created_at: new Date(Date.now() - 2 * 3600_000).toISOString(),
  title: 'Deploy api image',
  summary: 'New image tag deployed to api component.',
  app_branch: { id: 'branch_1', name: 'main', sha: 'b2c3d4e5f6g7' },
  workflow: {
    id: 'wf_2',
    name: 'Deploy workflow',
    type: 'component_deploy',
  },
  component_name: 'api',
  image: {
    repository: 'ghcr.io/acme/api',
    previous_tag: 'v1.2.3',
    next_tag: 'v1.3.0',
  },
  affected_resources: {
    components: ['api'],
    images: ['ghcr.io/acme/api'],
  },
  change_groups: [
    {
      id: 'cg_1',
      scope: 'component',
      label: 'API component changes',
      resource_name: 'api',
      summary: 'Image tag updated from v1.2.3 to v1.3.0.',
      changes: [
        {
          path: 'image.tag',
          operation: 'change',
          previous_value: 'v1.2.3',
          next_value: 'v1.3.0',
        },
      ],
    },
  ],
}

const DEPLOYMENT_CONFIG_UPDATE: TInstallDeploymentRecord = {
  id: 'dep_config_3',
  type: 'install_config_update',
  status: 'error',
  created_at: new Date(Date.now() - 48 * 3600_000).toISOString(),
  title: 'Config update',
  summary: 'Install config file synced.',
  app_branch: { id: 'branch_1', name: 'main' },
  affected_resources: {
    components: ['api', 'worker'],
    images: [],
  },
  change_groups: [
    {
      id: 'cg_2',
      scope: 'install_config',
      label: 'Config file changes',
      summary: 'Two fields updated.',
      changes: [
        {
          path: 'components.api.config.replicas',
          operation: 'change',
          previous_value: '2',
          next_value: '4',
        },
        {
          path: 'components.worker.image',
          operation: 'add',
          next_value: 'ghcr.io/acme/worker:v2.0.0',
        },
      ],
    },
  ],
}

const MOCK_DEPLOYMENTS = [
  DEPLOYMENT_IMAGE_UPDATE,
  DEPLOYMENT_PROVISION,
  DEPLOYMENT_CONFIG_UPDATE,
]

// ─── Stories ─────────────────────────────────────────────────────────────────

export const Default = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={MOCK_DEPLOYMENTS}
      isLoading={false}
      error={null}
      page={0}
      hasMore={false}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      filter={DEFAULT_DEPLOYMENTS_FILTER}
      onFilterChange={() => {}}
      onClearFilters={() => {}}
      onPageChange={() => {}}
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={[]}
      isLoading
      error={null}
      page={0}
      hasMore={false}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      filter={DEFAULT_DEPLOYMENTS_FILTER}
      onFilterChange={() => {}}
      onClearFilters={() => {}}
      onPageChange={() => {}}
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={[]}
      isLoading={false}
      error={null}
      page={0}
      hasMore={false}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      filter={DEFAULT_DEPLOYMENTS_FILTER}
      onFilterChange={() => {}}
      onClearFilters={() => {}}
      onPageChange={() => {}}
    />
  </div>
)

export const EmptyFiltered = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={[]}
      isLoading={false}
      error={null}
      page={0}
      hasMore={false}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      filter={{ ...DEFAULT_DEPLOYMENTS_FILTER, status: 'active' }}
      onFilterChange={() => {}}
      onClearFilters={() => {}}
      onPageChange={() => {}}
    />
  </div>
)

export const WithPagination = () => (
  <div className="max-w-3xl mx-auto p-6">
    <DeploymentsListPresenter
      deployments={MOCK_DEPLOYMENTS}
      isLoading={false}
      error={null}
      page={1}
      hasMore={true}
      orgId={ORG_ID}
      appId={APP_ID}
      installId={INSTALL_ID}
      filter={DEFAULT_DEPLOYMENTS_FILTER}
      onFilterChange={() => {}}
      onClearFilters={() => {}}
      onPageChange={() => {}}
    />
  </div>
)

export const DetailPanelStory = () => (
  <PanelStory label="Open deployment details">
    <DeploymentDetailPanel
      deployment={DEPLOYMENT_IMAGE_UPDATE}
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
