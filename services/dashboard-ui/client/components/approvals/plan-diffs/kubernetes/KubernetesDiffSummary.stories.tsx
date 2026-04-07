export default {
  title: 'Approvals/PlanDiffs/KubernetesDiffSummary',
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
