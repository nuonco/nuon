export default {
  title: 'Admin/AdminControls',
}

import { AdminControls } from './AdminControls'

export const OrgOnly = () => (
  <AdminControls
    isNuonEmployee={true}
    orgId="org-123"
  />
)

export const WithApp = () => (
  <AdminControls
    isNuonEmployee={true}
    orgId="org-123"
    appId="app-456"
  />
)

export const WithInstall = () => (
  <AdminControls
    isNuonEmployee={true}
    orgId="org-123"
    appId="app-456"
    installId="install-789"
  />
)

export const NonEmployee = () => (
  <AdminControls
    isNuonEmployee={false}
    orgId="org-123"
  />
)
