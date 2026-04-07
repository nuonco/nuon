export default {
  title: 'Approvals/PlanDiffs/TerraformSummary',
}

import { TerraformSummary } from './TerraformSummary'

export const Default = () => (
    <TerraformSummary
      summary={{
        create: 4,
        update: 2,
        delete: 1,
        replace: 0,
        read: 3,
        'no-op': 5,
      }}
    />
  )

export const AllZero = () => (
    <TerraformSummary
      summary={{
        create: 0,
        update: 0,
        delete: 0,
        replace: 0,
        read: 0,
        'no-op': 0,
      }}
    />
  )
