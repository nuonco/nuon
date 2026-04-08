export default {
  title: 'Orgs/RecentActivities',
}

import { RecentActivities } from './RecentActivities'
import type { IActivity } from './RecentActivities'

const mockActivities: IActivity[] = [
  {
    id: 'act-1',
    installName: 'production',
    installId: 'install-1',
    message: 'deployed successfully',
    status: 'active',
    created_at: new Date(Date.now() - 300000).toISOString(),
    duration: '2m 14s',
    triggeredBy: 'jane@example.com',
    href: '/org-1/installs/install-1',
  },
  {
    id: 'act-2',
    installName: 'staging',
    installId: 'install-2',
    message: 'deploy failed',
    status: 'error',
    created_at: new Date(Date.now() - 900000).toISOString(),
    triggeredBy: 'bob@example.com',
    href: '/org-1/installs/install-2',
  },
  {
    id: 'act-3',
    installName: 'demo',
    installId: 'install-3',
    message: 'provisioned successfully',
    status: 'active',
    created_at: new Date(Date.now() - 3600000).toISOString(),
    duration: '8m 42s',
    triggeredBy: 'alice@example.com',
  },
]

export const Default = () => (
  <RecentActivities
    activities={mockActivities}
    pagination={{ limit: 10, offset: 0, hasNext: false }}
  />
)

export const Empty = () => <RecentActivities activities={[]} />

export const WithPagination = () => (
  <RecentActivities
    activities={mockActivities}
    pagination={{ limit: 3, offset: 0, hasNext: true }}
  />
)
