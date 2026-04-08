export default {
  title: 'Deploys/DeployTimeline',
}

import { DeployTimeline } from './DeployTimeline'
import type { TDeploy } from '@/types'

const mockDeploy: TDeploy = {
  id: 'dep-123',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-15T10:35:00Z',
  status: 'active',
  status_v2: { status: 'active' },
  install_deploy_type: 'apply',
  created_by: { email: 'user@example.com' },
} as TDeploy

const mockTeardown: TDeploy = {
  ...mockDeploy,
  id: 'dep-456',
  created_at: '2024-01-14T10:30:00Z',
  updated_at: '2024-01-14T10:35:00Z',
  install_deploy_type: 'teardown',
} as TDeploy

const mockDrifted: TDeploy = {
  ...mockDeploy,
  id: 'dep-789',
  created_at: '2024-01-13T10:30:00Z',
  updated_at: '2024-01-13T10:35:00Z',
  status: 'drifted',
  status_v2: { status: 'drifted' },
} as TDeploy

export const Default = () => (
  <DeployTimeline
    deploys={[mockDeploy, mockTeardown, mockDrifted]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    installId="install-1"
    componentId="comp-1"
    componentName="API server"
    isLoading={false}
    error={null}
  />
)

export const Loading = () => (
  <DeployTimeline
    deploys={[]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    installId="install-1"
    componentId="comp-1"
    componentName="API server"
    isLoading={true}
    error={null}
  />
)

export const Empty = () => (
  <DeployTimeline
    deploys={[]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    installId="install-1"
    componentId="comp-1"
    componentName="API server"
    isLoading={false}
    error={null}
  />
)

export const WithPagination = () => (
  <DeployTimeline
    deploys={[mockDeploy]}
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
    orgId="org-1"
    installId="install-1"
    componentId="comp-1"
    componentName="API server"
    isLoading={false}
    error={null}
  />
)
