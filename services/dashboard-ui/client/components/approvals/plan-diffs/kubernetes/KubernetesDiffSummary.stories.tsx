export default {
  title: 'Features / Approvals / Plan diffs / Kubernetes diff summary',
}

import { KubernetesDiffSummary } from './KubernetesDiffSummary'

export const Default = () => (
    <KubernetesDiffSummary
      summary={{
        add: 2,
        change: 5,
        destroy: 0,
      }}
    />
  )
