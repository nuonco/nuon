export default {
  title: 'Admin/AdminInstallSection',
}

import { AdminInstallSection } from './AdminInstallSection'

export const Default = () => (
  <AdminInstallSection
    orgId="org-1"
    installId="install-abc123"
    adminEmail="admin@nuon.co"
    runner={{ id: 'runner-xyz789' } as any}
    runnerLoading={false}
  />
)

export const Loading = () => (
  <AdminInstallSection
    orgId="org-1"
    installId="install-abc123"
    adminEmail="admin@nuon.co"
    runner={undefined}
    runnerLoading={true}
  />
)
