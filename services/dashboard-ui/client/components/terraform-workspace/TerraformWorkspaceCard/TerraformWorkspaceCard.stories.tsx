export default {
  title: 'Terraform/TerraformWorkspaceCard',
}

import { DateTime } from 'luxon'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { TerraformWorkspaceLockBadge } from '@/components/terraform-workspace/TerraformWorkspaceLockBadge'
import { TerraformWorkspaceCard } from './TerraformWorkspaceCard'

const mockState = {
  values: {
    outputs: {
      vpc_id: { value: 'vpc-abc123' },
      region: { value: 'us-west-2' },
    },
    root_module: {
      resources: [
        {
          address: 'aws_vpc.main',
          name: 'main',
          type: 'aws_vpc',
          mode: 'managed',
          provider_name: 'registry.terraform.io/hashicorp/aws',
          schema_version: 1,
          values: { cidr_block: '10.0.0.0/16' },
          sensitive_values: {},
        },
      ],
    },
  },
} as any

const unlockButton = (
  <Button variant="danger" size="sm">
    <Icon variant="LockOpen" size={14} />
    Unlock Terraform state
  </Button>
)

const cliButton = (
  <Button variant="secondary" size="sm">Use Terraform CLI</Button>
)

const lockActions = (
  <>
    {cliButton}
    {unlockButton}
  </>
)

const mockLockBadge = (lock: Parameters<typeof TerraformWorkspaceLockBadge>[0]['lock']) => (
  <TerraformWorkspaceLockBadge lock={lock} />
)

export const Default = () => (
  <TerraformWorkspaceCard
    currentRevision={mockState}
    actions={lockActions}
  />
)

export const Empty = () => (
  <TerraformWorkspaceCard
    currentRevision={null}
    actions={cliButton}
  />
)

export const Unlocked = () => (
  <TerraformWorkspaceCard
    currentRevision={mockState}
    actions={cliButton}
  />
)

export const LockedByPlan = () => (
  <TerraformWorkspaceCard
    currentRevision={mockState}
    status={mockLockBadge({
      id: 'lock-1',
      created_at: DateTime.now().minus({ minutes: 3 }).toISO()!,
      lock: {
        operation: 'OperationTypePlan',
        who: 'runner-job-abc123',
        created: DateTime.now().minus({ minutes: 3 }).toISO()!,
      },
    })}
    actions={lockActions}
  />
)

export const LockedByApply = () => (
  <TerraformWorkspaceCard
    currentRevision={mockState}
    status={mockLockBadge({
      id: 'lock-2',
      created_at: DateTime.now().minus({ minutes: 12 }).toISO()!,
      lock: {
        operation: 'OperationTypeApply',
        who: 'runner-job-def456',
        created: DateTime.now().minus({ minutes: 12 }).toISO()!,
      },
    })}
    actions={lockActions}
  />
)

export const LockedByCLI = () => (
  <TerraformWorkspaceCard
    currentRevision={mockState}
    status={mockLockBadge({
      id: 'lock-3',
      created_at: DateTime.now().minus({ hours: 2 }).toISO()!,
      lock: {
        operation: 'OperationTypePlan',
        who: 'test@nuon.co',
        created: DateTime.now().minus({ hours: 2 }).toISO()!,
      },
    })}
    actions={lockActions}
  />
)

export const StaleLock = () => (
  <TerraformWorkspaceCard
    currentRevision={mockState}
    status={mockLockBadge({
      id: 'lock-4',
      created_at: DateTime.now().minus({ hours: 6 }).toISO()!,
      lock: {
        operation: 'OperationTypeApply',
        created: DateTime.now().minus({ hours: 6 }).toISO()!,
      },
    })}
    actions={lockActions}
  />
)

export const EmptyAndLocked = () => (
  <TerraformWorkspaceCard
    currentRevision={null}
    status={mockLockBadge({
      id: 'lock-5',
      created_at: DateTime.now().toISO()!,
      lock: {
        operation: 'OperationTypeInit',
        who: 'runner-job-ghi789',
        created: DateTime.now().toISO()!,
      },
    })}
    actions={lockActions}
  />
)
