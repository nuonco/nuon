export default {
  title: 'Install Components/InstallComponentDependencies',
}

import { InstallComponentDependencies } from './InstallComponentDependencies'

const mockComponents = [
  { id: 'comp-1', name: 'database' },
  { id: 'comp-2', name: 'cache' },
  { id: 'comp-3', name: 'queue' },
] as any[]

export const CountVariant = () => (
  <InstallComponentDependencies
    deps={['comp-1', 'comp-2']}
    variant="count"
    components={mockComponents.slice(0, 2)}
    isLoading={false}
    basePath="/org/installs/inst/components"
    pathname="/org/installs/inst/components"
  />
)

export const InlineVariant = () => (
  <InstallComponentDependencies
    deps={['comp-1', 'comp-2', 'comp-3']}
    variant="inline"
    components={mockComponents}
    isLoading={false}
    basePath="/org/installs/inst/components"
    pathname="/org/installs/inst/components"
  />
)

export const Loading = () => (
  <InstallComponentDependencies
    deps={['comp-1']}
    variant="count"
    components={[]}
    isLoading={true}
    basePath="/org/installs/inst/components"
    pathname="/org/installs/inst/components"
  />
)

export const Empty = () => (
  <InstallComponentDependencies
    deps={[]}
    variant="count"
    components={[]}
    isLoading={false}
    basePath="/org/installs/inst/components"
    pathname="/org/installs/inst/components"
  />
)
