export default {
  title: 'Admin/TemporalLink',
}

import { TemporalLink } from './TemporalLink'

export const Visible = () => (
  <TemporalLink
    namespace="components"
    eventLoopId="wf-123"
    isVisible={true}
  />
)

export const Hidden = () => (
  <TemporalLink
    namespace="components"
    eventLoopId="wf-123"
    isVisible={false}
  />
)
