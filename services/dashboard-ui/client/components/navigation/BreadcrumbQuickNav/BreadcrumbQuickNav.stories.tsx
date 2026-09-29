import { BreadcrumbQuickNav } from './BreadcrumbQuickNav'

export default {
  title: 'UI / Navigation / Breadcrumb quick nav',
}

const items = [
  { id: 'branch-1', name: 'main', href: '#' },
  { id: 'branch-2', name: 'staging', href: '#' },
  { id: 'branch-3', name: 'feature/payments', href: '#' },
]

export const Default = () => (
  <BreadcrumbQuickNav
    id="breadcrumb-quick-nav-story"
    label="main"
    title="Switch branch"
    items={items}
    currentId="branch-1"
    isLoading={false}
    searchTerm=""
    onSearch={() => {}}
  />
)

export const Loading = () => (
  <BreadcrumbQuickNav
    id="breadcrumb-quick-nav-loading-story"
    label="production"
    title="Switch install"
    items={[]}
    currentId="install-1"
    isLoading
    searchTerm=""
    onSearch={() => {}}
  />
)
