export default {
  title: 'Admin/TemporalLink',
}

import { TemporalLink } from './TemporalLink'

export const Visible = () => (
  <TemporalLink
    href="/admin/temporal/namespaces/components/workflows/event-loop-wf-123"
    isVisible={true}
  />
)

export const Hidden = () => (
  <TemporalLink
    href="/admin/temporal/namespaces/components/workflows/event-loop-wf-123"
    isVisible={false}
  />
)
