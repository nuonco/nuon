export default {
  title: 'Workflows/WorkflowActionButtons',
}

import { AuthContext } from '@/providers/auth-provider'
import { ConfigContext, type TRuntimeConfig } from '@/providers/config-provider'
import { WorkflowActionButtons } from './WorkflowActionButtons'
import type { TWorkflow } from '@/types'

const mockWorkflow = {
  id: 'wf-123',
  owner_id: 'inst-456',
  type: 'deploy_components',
  status: { status: 'in-progress' },
  finished: false,
  approval_option: 'prompt',
} as TWorkflow

const mockConfig: TRuntimeConfig = {
  apiUrl: '',
  appUrl: '',
  githubAppName: '',
  isByoc: false,
  adminDashboardUrl: 'https://admin.example.com',
}

const adminAuth = {
  user: { sub: '1', email: 'test@nuon.co' },
  isAuthenticated: true,
  isAdmin: true,
  isNuonEmployee: true,
  isLoading: false,
  error: null,
  demoMode: false,
  toggleDemoMode: () => {},
}

const nonAdminAuth = {
  ...adminAuth,
  isAdmin: false,
}

export const AllButtons = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={adminAuth}>
      <WorkflowActionButtons
        workflow={mockWorkflow}
        canShowApproveAll={true}
        canShowCancel={true}
      />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)

export const CancelOnly = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={adminAuth}>
      <WorkflowActionButtons
        workflow={mockWorkflow}
        canShowApproveAll={false}
        canShowCancel={true}
      />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)

export const Empty = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={nonAdminAuth}>
      <WorkflowActionButtons
        workflow={mockWorkflow}
        canShowApproveAll={false}
        canShowCancel={false}
      />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)
