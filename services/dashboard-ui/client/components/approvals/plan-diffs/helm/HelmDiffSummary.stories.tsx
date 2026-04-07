export default {
  title: 'Approvals/PlanDiffs/HelmDiffSummary',
}

import { HelmDiffSummary } from './HelmDiffSummary'

export const Default = () => (
    <HelmDiffSummary
      summary={{
        add: 3,
        change: 2,
        destroy: 1,
      }}
    />
  )

export const AllZero = () => (
    <HelmDiffSummary
      summary={{
        add: 0,
        change: 0,
        destroy: 0,
      }}
    />
  )
