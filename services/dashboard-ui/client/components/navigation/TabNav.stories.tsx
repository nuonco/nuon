export default {
  title: 'Navigation/TabNav',
}

import { MemoryRouter } from 'react-router'
import { TabNav } from './TabNav'

const mockTabs = [
  { path: '/', text: 'Overview', iconVariant: undefined },
  { path: '/deploys', text: 'Deploys', iconVariant: undefined },
  { path: '/actions', text: 'Actions', iconVariant: undefined },
  { path: '/logs', text: 'Logs', iconVariant: undefined },
]

export const Default = () => (
  <MemoryRouter initialEntries={['/org-1/installs/install-1']}>
    <TabNav basePath="/org-1/installs/install-1" tabs={mockTabs} />
  </MemoryRouter>
)

export const ActiveTab = () => (
  <MemoryRouter initialEntries={['/org-1/installs/install-1/deploys']}>
    <TabNav basePath="/org-1/installs/install-1" tabs={mockTabs} />
  </MemoryRouter>
)
