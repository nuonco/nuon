export default {
  title: 'Terraform/TerraformWorkspaceLockBadge',
}

import { DateTime } from 'luxon'
import { TerraformWorkspaceLockBadge } from './TerraformWorkspaceLockBadge'

export const LockedByPlan = () => (
  <TerraformWorkspaceLockBadge
    lock={{
      id: 'lock-1',
      created_at: DateTime.now().minus({ minutes: 3 }).toISO()!,
      lock: {
        operation: 'OperationTypePlan',
        who: 'test@nuon.co',
        created: DateTime.now().minus({ minutes: 3 }).toISO()!,
      },
    }}
  />
)

export const LockedByApply = () => (
  <TerraformWorkspaceLockBadge
    lock={{
      id: 'lock-2',
      created_at: DateTime.now().minus({ minutes: 12 }).toISO()!,
      lock: {
        operation: 'OperationTypeApply',
        who: 'runner-job-abc123',
        created: DateTime.now().minus({ minutes: 12 }).toISO()!,
      },
    }}
  />
)

export const LockedByCLI = () => (
  <TerraformWorkspaceLockBadge
    lock={{
      id: 'lock-3',
      created_at: DateTime.now().minus({ hours: 2 }).toISO()!,
      lock: {
        operation: 'OperationTypePlan',
        who: 'test@nuon.co',
        created: DateTime.now().minus({ hours: 2 }).toISO()!,
      },
    }}
  />
)

export const StaleLock = () => (
  <TerraformWorkspaceLockBadge
    lock={{
      id: 'lock-4',
      created_at: DateTime.now().minus({ hours: 6 }).toISO()!,
      lock: {
        operation: 'OperationTypeApply',
        created: DateTime.now().minus({ hours: 6 }).toISO()!,
      },
    }}
  />
)

export const JustNow = () => (
  <TerraformWorkspaceLockBadge
    lock={{
      id: 'lock-5',
      created_at: DateTime.now().toISO()!,
      lock: {
        operation: 'OperationTypePlan',
        who: 'runner-job-xyz789',
        created: DateTime.now().toISO()!,
      },
    }}
  />
)
