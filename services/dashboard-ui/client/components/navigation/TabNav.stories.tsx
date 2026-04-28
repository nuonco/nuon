export default {
  title: 'Navigation/TabNav',
}

import { TabNav } from './TabNav'

const mockTabs = [
  { path: '/', text: 'Overview' },
  { path: '/deploys', text: 'Deploys' },
  { path: '/actions', text: 'Actions' },
  { path: '/logs', text: 'Logs' },
]

export const Default = () => (
  <TabNav basePath="/org-1/installs/install-1" tabs={mockTabs} activeIndex={0} />
)

export const ActiveTab = () => (
  <TabNav basePath="/org-1/installs/install-1" tabs={mockTabs} activeIndex={2} />
)

const mockTabsWithBadge = [
  { path: '/', text: 'Logs' },
  { path: '/plan', text: 'Plan', badge: true },
  { path: '/variables', text: 'Variables' },
  { path: '/state', text: 'State' },
]

export const WithBadge = () => (
  <TabNav basePath="/org-1/installs/install-1/components/c-1/deploys/d-1" tabs={mockTabsWithBadge} activeIndex={0} />
)
