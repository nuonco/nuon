import { stepStatusCategory } from '@/components/branches/shared/step-status'
import { pluralize } from '@/components/branches/BranchOverview/rollout-stages'
import type { TWorkflow } from '@/types'

const DEPLOY_GROUP = /^deploy install group:\s*(.+)$/i

const installStatuses = (metadata?: Record<string, unknown>) => {
  const installs = metadata?.installs
  if (!Array.isArray(installs)) return []
  return installs.flatMap((install) => {
    if (!install || typeof install !== 'object') return []
    const status = (install as { status?: unknown }).status
    return typeof status === 'string' ? [status] : []
  })
}

export const branchRunOutcome = (
  workflow: TWorkflow
): { text: string; failed: boolean } | undefined => {
  const groups = (workflow.steps ?? []).flatMap((step) => {
    const name = step.name?.match(DEPLOY_GROUP)?.[1]?.trim()
    if (!name) return []
    return [
      {
        name,
        stepStatus: step.status?.status,
        installs: installStatuses(
          step.status?.metadata as Record<string, unknown> | undefined
        ),
      },
    ]
  })
  if (groups.length === 0) return undefined

  const failed = groups.find(
    (group) =>
      stepStatusCategory(group.stepStatus) === 'error' ||
      group.installs.some((status) => stepStatusCategory(status) === 'error')
  )
  if (failed) return { text: `Failed in ${failed.name}`, failed: true }

  const total = groups.reduce((sum, group) => sum + group.installs.length, 0)
  if (total === 0) return undefined
  const updated = groups.reduce(
    (sum, group) =>
      sum +
      group.installs.filter((status) => stepStatusCategory(status) === 'success')
        .length,
    0
  )
  if (updated === total) {
    return { text: `${pluralize(updated, 'install')} updated`, failed: false }
  }
  return {
    text: `${updated} of ${pluralize(total, 'install')} updated`,
    failed: false,
  }
}
