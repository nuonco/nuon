export default {
  title: 'VCS Connections/VCSConnectionsStatusIndicator',
}

import { Status } from '@/components/common/Status'
import { VCSConnectionsStatusIndicator } from './VCSConnectionsStatusIndicator'

const makeItems = (connections: Array<{ id: string; name: string; status: string }>) =>
  connections.map((conn) => ({
    id: conn.id,
    href: `/org-123/connections/vcs/${conn.id}`,
    title: conn.name,
    subtitle: conn.status,
    leftContent: (
      <Status
        status={conn.status === 'active' ? 'success' : conn.status === 'suspended' ? 'error' : 'warn'}
        isWithoutText
        variant="timeline"
        iconSize={16}
      />
    ),
  }))

export const AllActive = () => (
  <VCSConnectionsStatusIndicator
    theme="success"
    items={makeItems([
      { id: '1', name: 'nuonco', status: 'active' },
      { id: '2', name: 'nuon-internal', status: 'active' },
    ])}
  />
)

export const OneWarn = () => (
  <VCSConnectionsStatusIndicator
    theme="warn"
    items={makeItems([
      { id: '1', name: 'nuonco', status: 'active' },
      { id: '2', name: 'nuon-internal', status: 'unknown' },
    ])}
  />
)

export const OneSuspended = () => (
  <VCSConnectionsStatusIndicator
    theme="error"
    items={makeItems([
      { id: '1', name: 'nuonco', status: 'active' },
      { id: '2', name: 'nuon-internal', status: 'suspended' },
    ])}
  />
)

export const Single = () => (
  <VCSConnectionsStatusIndicator
    theme="success"
    items={makeItems([{ id: '1', name: 'nuonco', status: 'active' }])}
  />
)

export const Loading = () => (
  <VCSConnectionsStatusIndicator
    theme="neutral"
    items={makeItems([{ id: '1', name: 'nuonco', status: 'unknown' }])}
  />
)
