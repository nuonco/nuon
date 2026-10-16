export const BRANCH_CONFIG_POLL_MS = 5_000

const failedRunStatuses = new Set(['failed', 'error', 'cancelled'])

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
}): { phase: 'skip' | 'ready' | 'waiting' | 'failed'; message?: string } => {
  if (!input.hasBranch) return { phase: 'skip' }
  if (input.hasActiveConfig) return { phase: 'ready' }
  const runs = input.runs ?? []
  const latest = runs[0]
  const status = latest?.status?.status ?? ''
  if (runs.length === 1 && failedRunStatuses.has(status)) {
    return {
      phase: 'failed',
      message:
        latest?.status?.status_human_description ||
        'This branch run failed before an app config was created.',
    }
  }
  return { phase: 'waiting' }
}
