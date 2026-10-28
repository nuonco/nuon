export default {
  title: 'Playground / Customer portal / Shared with',
}

import { mockShares } from '../fixtures'
import { SharedWithList } from './SharedWithList'

const noop = () => {}

export const Empty = () => (
  <div className="p-4 max-w-3xl">
    <SharedWithList shares={[]} onRevoke={noop} onResend={noop} />
  </div>
)

export const Populated = () => (
  <div className="p-4 max-w-3xl">
    <SharedWithList
      shares={mockShares.filter((share) => share.install_id === 'inst-1')}
      onRevoke={noop}
      onResend={noop}
    />
  </div>
)

export const Loading = () => (
  <div className="p-4 max-w-3xl">
    <SharedWithList shares={[]} isLoading onRevoke={noop} onResend={noop} />
  </div>
)

export const ErrorState = () => (
  <div className="p-4 max-w-3xl">
    <SharedWithList shares={[]} isError onRevoke={noop} onResend={noop} />
  </div>
)
