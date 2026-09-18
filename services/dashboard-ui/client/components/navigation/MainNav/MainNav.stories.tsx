export default {
  title: 'Navigation/MainNav',
}

import { MainNav } from './MainNav'

const mockOrg = {
  id: 'org-1',
  name: 'My Org',
  features: {},
} as any

export const Default = () => (
  <div className="w-[248px] p-4">
    <MainNav
      org={mockOrg}
      isSidebarOpen
      showInstalls
      hasCustomerPortal={false}
      customerPortalUrl="https://customers.nuon.co"
    />
  </div>
)

export const Collapsed = () => (
  <div className="w-[60px] p-4">
    <MainNav
      org={mockOrg}
      isSidebarOpen={false}
      showInstalls
      hasCustomerPortal={false}
      customerPortalUrl="https://customers.nuon.co"
    />
  </div>
)

export const WithoutInstalls = () => (
  <div className="w-[248px] p-4">
    <MainNav
      org={mockOrg}
      isSidebarOpen
      showInstalls={false}
      hasCustomerPortal={false}
      customerPortalUrl="https://customers.nuon.co"
    />
  </div>
)
