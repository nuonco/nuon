export const BRANCH_CONFIG_POLL_MS = 5_000

const failedRunStatuses = new Set(['failed', 'error', 'cancelled', 'discarded'])
const inFlightRunStatuses = new Set([
  'pending',
  'queued',
  'in-progress',
  'planning',
  'applying',
  'checking-plan',
  'retrying',
  'failed-pending-retry',
])

export interface IBranchConfigRun {
  status?: {
    status?: string
    status_human_description?: string
  }
}

export const branchConfigWaitState = (input: {
  hasBranch: boolean
  hasActiveConfig: boolean
  runs?: IBranchConfigRun[]
  latestConfigStatus?: string
}): { phase: 'skip' | 'ready' | 'waiting' | 'failed'; message?: string } => {
  if (!input.hasBranch) return { phase: 'skip' }
  if (input.hasActiveConfig) return { phase: 'ready' }
  const runs = input.runs ?? []
  const latest = runs[0]
  const status = latest?.status?.status ?? ''

  if (inFlightRunStatuses.has(status)) return { phase: 'waiting' }

  if (failedRunStatuses.has(status)) {
    return {
      phase: 'failed',
      message:
        latest?.status?.status_human_description ||
        'This branch run failed before an app config was created.',
    }
  }

  if (status === 'not-attempted') {
    return {
      phase: 'failed',
      message:
        'The latest branch run was skipped, so this branch has no app config. Run the branch again.',
    }
  }

  if (status === 'success' && input.latestConfigStatus === 'error') {
    return {
      phase: 'failed',
      message: "This branch's app config failed to build.",
    }
  }

  return { phase: 'waiting' }
}
