export default {
  title: 'Components/ComponentDependencies',
}

import { ComponentDependencies } from './ComponentDependencies'

export const CountVariant = () => (
  <ComponentDependencies
    deps={['dep-1', 'dep-2']}
    variant="count"
    components={[
      { id: 'dep-1', name: 'Component A' } as any,
      { id: 'dep-2', name: 'Component B' } as any,
    ]}
    isLoading={false}
    basePath="/org-1/apps/app-1/components"
    pathname="/org-1/apps/app-1/components/comp-1"
  />
)

export const InlineVariant = () => (
  <ComponentDependencies
    deps={['dep-1', 'dep-2', 'dep-3']}
    variant="inline"
    components={[
      { id: 'dep-1', name: 'Component A' } as any,
      { id: 'dep-2', name: 'Component B' } as any,
      { id: 'dep-3', name: 'Component C' } as any,
    ]}
    isLoading={false}
    basePath="/org-1/apps/app-1/components"
    pathname="/org-1/apps/app-1/components/comp-1"
  />
)

export const Loading = () => (
  <ComponentDependencies
    deps={['dep-1']}
    variant="count"
    components={[]}
    isLoading={true}
    basePath="/org-1/apps/app-1/components"
    pathname="/org-1/apps/app-1/components/comp-1"
  />
)

export const Empty = () => (
  <ComponentDependencies
    deps={[]}
    variant="count"
    components={[]}
    isLoading={false}
    basePath="/org-1/apps/app-1/components"
    pathname="/org-1/apps/app-1/components/comp-1"
  />
)
