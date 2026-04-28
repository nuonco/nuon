export default {
  title: 'Runners/RunnerProcessesTable',
}

import { AuthContext } from '@/providers/auth-provider'
import { ConfigContext, type TRuntimeConfig } from '@/providers/config-provider'
import { RunnerProcessesTable } from './RunnerProcessesTable'

const mockConfig: TRuntimeConfig = {
  apiUrl: '',
  appUrl: '',
  githubAppName: '',
  isByoc: false,
  adminDashboardUrl: 'http://localhost:8085',
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

const mockProcesses = [
  { id: 'proc-1', type: 'build', composite_status: { status: 'active' }, version: 'development', labels: ['Local Runner'], started_at: new Date(Date.now() - 24 * 60 * 1000).toISOString(), created_at: new Date(Date.now() - 24 * 60 * 1000).toISOString(), runner_id: 'runner-1' },
  { id: 'proc-2', type: 'build', composite_status: { status: 'inactive' }, version: 'development', labels: ['Local Runner'], started_at: new Date(Date.now() - 52 * 60 * 1000).toISOString(), created_at: new Date(Date.now() - 52 * 60 * 1000).toISOString(), runner_id: 'runner-1' },
] as any[]

export const Default = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={adminAuth}>
      <RunnerProcessesTable processes={mockProcesses} isLoading={false} />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)

export const NonAdmin = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={nonAdminAuth}>
      <RunnerProcessesTable processes={mockProcesses} isLoading={false} />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)

export const Loading = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={adminAuth}>
      <RunnerProcessesTable processes={[]} isLoading={true} />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)

export const Empty = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={adminAuth}>
      <RunnerProcessesTable processes={[]} isLoading={false} />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)
