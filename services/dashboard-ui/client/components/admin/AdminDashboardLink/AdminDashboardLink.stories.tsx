import { AuthContext } from '@/providers/auth-provider'
import { ConfigContext, type TRuntimeConfig } from '@/providers/config-provider'
import { AdminDashboardLink } from './AdminDashboardLink'

export default {
  title: 'Admin/AdminDashboardLink',
}

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

export const Default = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={adminAuth}>
      <AdminDashboardLink
        path="/queues?owner_id=inst123&owner_type=installs"
        label="View queues"
      />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)

export const Hidden = () => (
  <ConfigContext.Provider value={mockConfig}>
    <AuthContext.Provider value={nonAdminAuth}>
      <AdminDashboardLink
        path="/queues"
        label="View queues"
      />
    </AuthContext.Provider>
  </ConfigContext.Provider>
)
