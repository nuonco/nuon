export default {
  title: 'Navigation/SubNavLink',
}

import { SubNavLink } from './SubNavLink'

export const Default = () => (
  <nav className="flex flex-col gap-1 p-4 w-[280px]">
    <SubNavLink
      basePath="/org-123/installs/install-456"
      path="overview"
      text="Overview"
      iconVariant="HouseSimple"
    />
    <SubNavLink
      basePath="/org-123/installs/install-456"
      path="components"
      text="Components"
      iconVariant="Stack"
    />
    <SubNavLink
      basePath="/org-123/installs/install-456"
      path="deploys"
      text="Deploys"
      iconVariant="ShippingContainer"
    />
  </nav>
)

export const NoIcon = () => (
  <nav className="flex flex-col gap-1 p-4 w-[280px]">
    <SubNavLink
      basePath="/org-123/installs/install-456"
      path="overview"
      text="Overview"
    />
    <SubNavLink
      basePath="/org-123/installs/install-456"
      path="settings"
      text="Settings"
    />
  </nav>
)
