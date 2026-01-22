'use client'

import { Status } from '@/components/common/Status'
import { Tooltip } from '@/components/common/Tooltip'
import { Text } from '@/components/common/Text'
import { useTerraformWorkspaceLock } from '@/hooks/useTerraformWorkspaceLock'

interface ITerraformLockStatusBadge {
  workspaceId?: string
}

export const TerraformLockStatusBadge = ({
  workspaceId,
}: ITerraformLockStatusBadge) => {
  const { lockStatus, isLocked } = useTerraformWorkspaceLock(workspaceId)

  if (!workspaceId) {
    return null
  }

  const badge = (
    <Status variant="badge" status={isLocked ? 'locked' : 'unlocked'} />
  )

  // Show tooltip with details when locked
  if (isLocked && lockStatus) {
    const lockedAt = lockStatus.created_at
      ? new Date(lockStatus.created_at).toLocaleString()
      : 'Unknown'

    return (
      <Tooltip
        position="top"
        tipContent={
          <div className="flex flex-col gap-1 p-2">
            <Text variant="subtext" weight="strong">
              Workspace Locked
            </Text>
            <Text variant="subtext">
              Job: {lockStatus.runner_job_id || 'unknown'}
            </Text>
            <Text variant="subtext">Locked: {lockedAt}</Text>
          </div>
        }
      >
        {badge}
      </Tooltip>
    )
  }

  return badge
}
