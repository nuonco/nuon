'use client'

import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { useTerraformWorkspaceLock } from '@/hooks/useTerraformWorkspaceLock'
import type { TRunnerJob, TRunnerJobStatus } from '@/types/ctl-api.types'

interface ITerraformLockBanner {
  workspaceId?: string
  sandboxRuns?: Array<{
    runner_jobs?: Array<TRunnerJob>
  }>
}

// Check if any job is currently in-progress
const hasAnyInProgressJob = (
  sandboxRuns?: Array<{
    runner_jobs?: Array<TRunnerJob>
  }>
): boolean => {
  if (!sandboxRuns || sandboxRuns.length === 0) return false

  const inProgressStatus: TRunnerJobStatus = 'in-progress'

  return sandboxRuns.some((run) =>
    run.runner_jobs?.some((job) => job.status === inProgressStatus)
  )
}

export const TerraformLockBanner = ({
  workspaceId,
  sandboxRuns,
}: ITerraformLockBanner) => {
  const { lockStatus, isLocked } = useTerraformWorkspaceLock(workspaceId)

  // Only show banner when locked AND no in-progress jobs (stale lock)
  if (!isLocked || !lockStatus) {
    return null
  }

  // Don't show banner if any job is currently in-progress
  if (hasAnyInProgressJob(sandboxRuns)) {
    return null
  }

  const lockedAt = lockStatus.created_at
    ? new Date(lockStatus.created_at).toLocaleString()
    : 'Unknown'

  return (
    <Banner theme="warn">
      <div className="flex flex-col gap-1">
        <Text weight="strong">Workspace Locked</Text>
        <Text variant="caption">
          This workspace is currently locked by job{' '}
          <code className="text-xs bg-orange-100 dark:bg-orange-900/20 px-1 py-0.5 rounded">
            {lockStatus.runner_job_id || 'unknown'}
          </code>{' '}
          (locked at {lockedAt}). Operations cannot proceed until the lock is
          released.
        </Text>
      </div>
    </Banner>
  )
}
