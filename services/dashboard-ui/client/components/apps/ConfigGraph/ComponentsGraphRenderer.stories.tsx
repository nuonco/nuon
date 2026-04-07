export default {
  title: 'Apps/ComponentsGraphRenderer',
}

import { ComponentsGraphInline } from './ComponentsGraphRenderer'

export const Loading = () => (
  <ComponentsGraphInline isLoading={true} />
)

export const Error = () => (
  <ComponentsGraphInline
    isLoading={false}
    error={{ error: 'Unable to load component change graph.' } as any}
  />
)

export const Empty = () => (
  <ComponentsGraphInline isLoading={false} dotGraph="" />
)
