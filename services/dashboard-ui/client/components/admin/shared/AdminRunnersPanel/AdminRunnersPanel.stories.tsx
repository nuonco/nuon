export default {
  title: 'Admin/AdminRunnersPanel',
}

import { AdminRunnersPanel } from './AdminRunnersPanel'

export const Default = () => (
  <AdminRunnersPanel
    orgId="org-1"
    orgName="My Org"
    orgRunners={[]}
    installs={[]}
    isLoading={false}
    isRestarting={false}
    onRestartAll={() => {}}
    onRefreshInstalls={() => {}}
  />
)
