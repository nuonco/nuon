export default {
  title: 'Navigation/MainNavLink',
}

import { MainNavLink } from './MainNavLink'

export const Default = () => (
  <MainNavLink
    basePath="/org-123"
    path="/installs"
    text="Installs"
    iconVariant="ShippingContainerIcon"
  />
)

export const Active = () => (
  <MainNavLink
    basePath=""
    path="/"
    text="Dashboard"
    iconVariant="HouseSimpleIcon"
  />
)

export const External = () => (
  <MainNavLink
    basePath="/org-123"
    path="https://docs.nuon.co"
    text="Docs"
    iconVariant="BookOpenIcon"
    isExternal
  />
)

export const NoIcon = () => (
  <MainNavLink
    basePath="/org-123"
    path="/settings"
    text="Settings"
  />
)
