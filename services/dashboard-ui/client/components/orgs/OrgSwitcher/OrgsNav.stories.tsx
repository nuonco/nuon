export default {
  title: 'Orgs/OrgsNav',
}

import { OrgsNav } from './OrgsNav'

const mockOrgs = Array.from({ length: 3 }, (_, i) => ({
  id: `org-${i + 1}`,
  name: `Organization ${i + 1}`,
})) as any[]

export const Default = () => (
  <OrgsNav
    orgs={mockOrgs}
    isLoading={false}
    searchTerm=""
    onSearchChange={() => {}}
    onLoadMore={() => {}}
    showSearch={false}
    showLoadMore={false}
  />
)

export const Loading = () => (
  <OrgsNav
    orgs={undefined}
    isLoading={true}
    searchTerm=""
    onSearchChange={() => {}}
    onLoadMore={() => {}}
    showSearch={false}
    showLoadMore={false}
  />
)

export const Empty = () => (
  <OrgsNav
    orgs={[]}
    isLoading={false}
    searchTerm="nonexistent"
    onSearchChange={() => {}}
    onLoadMore={() => {}}
    showSearch={true}
    showLoadMore={false}
  />
)
