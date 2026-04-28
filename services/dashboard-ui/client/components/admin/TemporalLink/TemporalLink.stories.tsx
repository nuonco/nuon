import { AuthContext } from '@/providers/auth-provider'
import { TemporalLink } from './TemporalLink'

export default {
  title: 'Admin/TemporalLink',
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

export const Visible = () => (
  <AuthContext.Provider value={adminAuth}>
    <TemporalLink
      namespace="components"
      eventLoopId="wf-123"
    />
  </AuthContext.Provider>
)

export const WithHref = () => (
  <AuthContext.Provider value={adminAuth}>
    <TemporalLink
      namespace=""
      href="/admin/temporal/namespaces/components/workflows/event-loop-wf-123"
    />
  </AuthContext.Provider>
)

export const Hidden = () => (
  <AuthContext.Provider value={nonAdminAuth}>
    <TemporalLink
      namespace="components"
      eventLoopId="wf-123"
    />
  </AuthContext.Provider>
)
