export default {
  title: 'Navigation/TabNav',
}

import { TabNav } from './TabNav'

const mockTabs = [
  { path: '/', text: 'Overview', iconVariant: undefined },
  { path: '/deploys', text: 'Deploys', iconVariant: undefined },
  { path: '/actions', text: 'Actions', iconVariant: undefined },
  { path: '/logs', text: 'Logs', iconVariant: undefined },
]

export const Default = () => (
  <TabNav basePath="/org-1/installs/install-1" tabs={mockTabs} />
)

export const ActiveTab = () => (
  <TabNav basePath="/org-1/installs/install-1" tabs={mockTabs} />
)

const mockTabsWithBadge = [
  { path: '/', text: 'Logs' },
  { path: '/plan', text: 'Plan', badge: true },
  { path: '/variables', text: 'Variables' },
  { path: '/state', text: 'State' },
]

export const WithBadge = () => (
  <TabNav basePath="/org-1/installs/install-1/components/c-1/deploys/d-1" tabs={mockTabsWithBadge} />
)
