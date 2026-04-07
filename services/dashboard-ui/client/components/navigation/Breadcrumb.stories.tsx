export default {
  title: 'Navigation/BreadcrumbNav',
}

import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { BreadcrumbNav } from './Breadcrumb'

const noop = () => {}

export const Default = () => (
  <BreadcrumbContext.Provider
    value={{
      breadcrumbLinks: [
        { path: '/org-001', text: 'My org' },
        { path: '/org-001/installs', text: 'Installs' },
        { path: '/org-001/installs/inst-001', text: 'production-install' },
      ],
      isLoading: false,
      updateBreadcrumb: noop,
    }}
  >
    <BreadcrumbNav />
  </BreadcrumbContext.Provider>
)

export const SingleCrumb = () => (
  <BreadcrumbContext.Provider
    value={{
      breadcrumbLinks: [{ path: '/org-001', text: 'My org' }],
      isLoading: false,
      updateBreadcrumb: noop,
    }}
  >
    <BreadcrumbNav />
  </BreadcrumbContext.Provider>
)

export const Loading = () => (
  <BreadcrumbContext.Provider
    value={{
      breadcrumbLinks: [
        { path: '/org-001', text: 'My org' },
        { path: '/org-001/installs', text: 'Installs' },
        { path: '/org-001/installs/inst-001', text: 'production-install' },
      ],
      isLoading: true,
      updateBreadcrumb: noop,
    }}
  >
    <BreadcrumbNav />
  </BreadcrumbContext.Provider>
)
