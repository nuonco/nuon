import type {
  TAppBranchInstallGroup,
  TInstall,
  TInstallWorkflowStep,
} from '@/types'
import {
  getStepDisplayStatus,
  stepStatusCategory,
  type TStepStatusCategory,
} from '@/components/branches/shared/step-status'

export type TRolloutStageKind = 'commit' | 'build' | 'group'

export interface TRolloutInstall {
  id: string
  name: string
  status: string
  region?: string
}

export interface TRolloutStage {
  id: string
  kind: TRolloutStageKind
  name: string
  status: string
  detail?: string
  durationNs?: number
  installs?: TRolloutInstall[]
  stepId?: string
  groupId?: string
}

const SKIPPED = new Set([
  'auto-skipped',
  'user-skipped',
  'not-attempted',
  'noop',
])

const CATEGORY_RANK: Record<TStepStatusCategory, number> = {
  error: 4,
  awaiting: 3,
  active: 2,
  pending: 1,
  success: 0,
}

const CATEGORY_STATUS: Record<TStepStatusCategory, string> = {
  error: 'error',
  awaiting: 'approval-awaiting',
  active: 'in-progress',
  pending: 'pending',
  success: 'success',
}

const GROUP_STEP = /^(?:plan|deploy) install group:\s*(.+)$/i
const RUNBOOK_STEP = /^run post-deploy runbooks:\s*(.+)$/i

export const pluralize = (count: number, noun: string) =>
  `${count} ${noun}${count === 1 ? '' : 's'}`

const isCommitStep = (name?: string) =>
  !!name && /commit/i.test(name) && !/status/i.test(name)

const isBuildStep = (name?: string) => !!name && /build/i.test(name)

const groupNameFromStep = (name?: string) => {
  if (!name) return undefined
  return (name.match(GROUP_STEP) ?? name.match(RUNBOOK_STEP))?.[1]?.trim()
}

const sameGroup = (stepName: string | undefined, groupName: string) => {
  const extracted = groupNameFromStep(stepName)
  return !!extracted && extracted.toLowerCase() === groupName.toLowerCase()
}

const normalizeInstallStatus = (status?: string) => {
  if (!status || status === 'deployed' || status === 'succeeded') {
    return status ? 'success' : 'pending'
  }
  return status
}

const aggregateStatus = (steps: TInstallWorkflowStep[]) => {
  if (steps.length === 0) return 'pending'
  const statuses = steps.map((step) => getStepDisplayStatus(step))
  if (statuses.every((status) => SKIPPED.has(status))) return 'auto-skipped'

  const meaningful = statuses.filter((status) => !SKIPPED.has(status))
  let best: TStepStatusCategory = 'success'
  for (const status of meaningful) {
    const category = stepStatusCategory(status)
    if (CATEGORY_RANK[category] > CATEGORY_RANK[best]) best = category
  }
  // A finished plan with a deploy still queued has started, so it stays in progress.
  if (
    best === 'pending' &&
    meaningful.some((status) => stepStatusCategory(status) === 'success')
  ) {
    return 'in-progress'
  }
  return CATEGORY_STATUS[best]
}

const durationNs = (steps: TInstallWorkflowStep[]) => {
  if (steps.length === 0 || steps.some((step) => !step.finished))
    return undefined
  const total = steps.reduce((sum, step) => sum + (step.execution_time || 0), 0)
  return total || undefined
}

const deployEntries = (step?: TInstallWorkflowStep) => {
  const installs = step?.status?.metadata?.installs
  if (!Array.isArray(installs)) return []
  return installs as Array<{
    install_id?: string
    name?: string
    status?: string
  }>
}

const regionOf = (install?: TInstall) =>
  install?.aws_account?.region ||
  install?.gcp_account?.region ||
  install?.azure_account?.location

const stageFromStep = (
  kind: 'commit' | 'build',
  fallbackName: string,
  step: TInstallWorkflowStep | undefined,
  detail?: string
): TRolloutStage => ({
  id: step?.id || kind,
  kind,
  name: fallbackName,
  status: step ? getStepDisplayStatus(step) : 'pending',
  detail,
  durationNs: step?.finished ? step.execution_time || undefined : undefined,
  stepId: step?.id,
})

export const buildRolloutStages = ({
  groups,
  steps,
  installsByGroup,
  sha,
}: {
  groups: TAppBranchInstallGroup[]
  steps: TInstallWorkflowStep[]
  installsByGroup: TInstall[][]
  sha?: string
}): TRolloutStage[] => {
  const ordered = [...steps].sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0))
  const commit = ordered.find((step) => isCommitStep(step.name))
  const build = ordered.find((step) => isBuildStep(step.name))
  const orderedGroups = groups
    .map((group, index) => ({
      group,
      installs: installsByGroup[index] ?? [],
    }))
    .sort((a, b) => (a.group.order ?? 0) - (b.group.order ?? 0))

  return [
    stageFromStep('commit', 'Commit', commit, sha?.slice(0, 7)),
    stageFromStep('build', 'Build', build),
    ...orderedGroups.map(({ group, installs: known }, index) => {
      const name = group.name || 'Install group'
      const groupSteps = ordered.filter((step) => sameGroup(step.name, name))
      const deployStep = groupSteps.find((step) =>
        /^deploy install group:/i.test(step.name ?? '')
      )
      const entries = deployEntries(deployStep)
      const knownById = new Map(known.map((install) => [install.id, install]))
      const installs = entries.length
        ? entries.map((entry) => {
            const install = entry.install_id
              ? knownById.get(entry.install_id)
              : undefined
            return {
              id: entry.install_id || install?.id || entry.name || name,
              name:
                install?.name || entry.name || entry.install_id || 'Install',
              status: normalizeInstallStatus(entry.status),
              region: regionOf(install),
            }
          })
        : undefined
      const preferred =
        deployStep ??
        groupSteps.find((step) =>
          /^plan install group:/i.test(step.name ?? '')
        ) ??
        groupSteps[0]
      const installFailed = entries.some(
        (entry) =>
          stepStatusCategory(normalizeInstallStatus(entry.status)) === 'error'
      )

      return {
        id: group.id || `group-${index}`,
        kind: 'group' as const,
        name,
        status: installFailed ? 'error' : aggregateStatus(groupSteps),
        detail: installs ? undefined : pluralize(known.length, 'install'),
        durationNs: durationNs(groupSteps),
        installs,
        stepId: preferred?.id,
        groupId: group.id,
      }
    }),
  ]
}
