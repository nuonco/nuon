import type {
  TAppBranchInstallGroup,
  TInstall,
  TInstallGroupRun,
} from '@/types'
import { hasLabelSelector, matchesSelector } from './label-selector'
import type { TLabelSelector } from './label-selector'

export type TInstallGroupMembership =
  | 'all_installs'
  | 'install_ids'
  | 'label_selector'

export interface IDeploymentPlanInstall {
  id: string
  name?: string
  labels?: Record<string, string>
}

export interface IDeploymentPlanStage {
  id: string
  name: string
  order: number
  stage: number
  membership: TInstallGroupMembership
  selector?: TLabelSelector
  maxParallel?: number
  autoApproveOnPoliciesPassing?: boolean
  installs: IDeploymentPlanInstall[]
  totalInstalls: number
  completedInstalls?: number
  failedInstalls?: number
  status?: string
}

const membershipOf = (
  group: TAppBranchInstallGroup
): TInstallGroupMembership => {
  if (group.all_installs) return 'all_installs'
  if (hasLabelSelector(group.label_selector)) return 'label_selector'
  return 'install_ids'
}

const toPlanInstall = (
  install: Pick<TInstall, 'id' | 'name' | 'labels'> & { id: string }
): IDeploymentPlanInstall => ({
  id: install.id,
  name: install.name,
  labels: install.labels,
})

const membersForGroup = ({
  group,
  membership,
  installs,
  claimed,
}: {
  group: TAppBranchInstallGroup
  membership: TInstallGroupMembership
  installs: Array<Pick<TInstall, 'id' | 'name' | 'labels'> & { id: string }>
  claimed: Set<string>
}): IDeploymentPlanInstall[] => {
  if (membership === 'all_installs') {
    return installs
      .filter((install) => !claimed.has(install.id))
      .map(toPlanInstall)
  }

  if (membership === 'label_selector') {
    return installs
      .filter(
        (install) =>
          !claimed.has(install.id) &&
          matchesSelector(install.labels, group.label_selector)
      )
      .map(toPlanInstall)
  }

  return (group.install_ids ?? []).flatMap((id) => {
    if (!id || claimed.has(id)) return []
    const install = installs.find((candidate) => candidate.id === id)
    return [install ? toPlanInstall(install) : { id }]
  })
}

export const resolveDeploymentPlanStages = ({
  groups,
  installs,
  groupRuns,
}: {
  groups?: TAppBranchInstallGroup[] | null
  installs?: Array<Pick<TInstall, 'id' | 'name' | 'labels'>> | null
  groupRuns?: TInstallGroupRun[] | null
}): IDeploymentPlanStage[] => {
  const orderedGroups = [...(groups ?? [])]
    .map((group, index) => ({ group, index }))
    .sort(
      (a, b) =>
        (a.group.order ?? 0) - (b.group.order ?? 0) || a.index - b.index
    )
    .map(({ group }) => group)

  const knownInstalls = (installs ?? []).filter(
    (
      install
    ): install is Pick<TInstall, 'id' | 'name' | 'labels'> & {
      id: string
    } => Boolean(install.id)
  )

  const runsByGroupId = new Map<string, TInstallGroupRun>()
  for (const run of groupRuns ?? []) {
    if (!run.install_group_id) continue
    runsByGroupId.set(run.install_group_id, run)
  }

  const claimed = new Set<string>()
  return orderedGroups.map((group, index) => {
    const membership = membershipOf(group)
    const members = membersForGroup({
      group,
      membership,
      installs: knownInstalls,
      claimed,
    })
    for (const member of members) {
      claimed.add(member.id)
    }

    const run = group.id ? runsByGroupId.get(group.id) : undefined
    const stage: IDeploymentPlanStage = {
      id: group.id ?? '',
      name: group.name ?? '',
      order: group.order ?? 0,
      stage: index + 1,
      membership,
      installs: members,
      totalInstalls: run?.total_installs ?? members.length,
    }

    if (membership === 'label_selector' && group.label_selector) {
      stage.selector = group.label_selector
    }
    if (group.max_parallel !== undefined) {
      stage.maxParallel = group.max_parallel
    }
    if (group.auto_approve_on_policies_passing != null) {
      stage.autoApproveOnPoliciesPassing = group.auto_approve_on_policies_passing
    }
    if (run?.completed_installs !== undefined) {
      stage.completedInstalls = run.completed_installs
    }
    if (run?.failed_installs !== undefined) {
      stage.failedInstalls = run.failed_installs
    }
    if (run?.status?.status) {
      stage.status = run.status.status
    }

    return stage
  })
}
