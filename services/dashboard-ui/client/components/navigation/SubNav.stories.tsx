export default {
  title: 'Navigation/SubNav',
}

import { SubNav } from './SubNav'

const links = [
  { path: 'overview', text: 'Overview', iconVariant: 'HouseSimple' as const },
  { path: 'components', text: 'Components', iconVariant: 'Stack' as const },
  { path: 'deploys', text: 'Deploys', iconVariant: 'ShippingContainer' as const },
  { path: 'actions', text: 'Actions', iconVariant: 'SneakerMove' as const },
]

export const Default = () => (
  <div className="h-screen flex">
    <SubNav basePath="/org-123/installs/install-456" links={links} />
    <div className="p-8">Page content</div>
  </div>
)

export const MinimalLinks = () => (
  <div className="h-screen flex">
    <SubNav
      basePath="/org-123/installs/install-456"
      links={[
        { path: 'overview', text: 'Overview', iconVariant: 'HouseSimple' as const },
        { path: 'settings', text: 'Settings', iconVariant: 'Gear' as const },
      ]}
    />
    <div className="p-8">Page content</div>
  </div>
)
