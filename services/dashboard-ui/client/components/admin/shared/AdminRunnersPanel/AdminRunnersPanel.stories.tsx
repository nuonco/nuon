export default {
  title: 'Features / Admin / Shared / Runners panel',
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
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
  />
)
