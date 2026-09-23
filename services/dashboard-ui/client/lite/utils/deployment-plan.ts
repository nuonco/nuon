import type {
  TAppBranchInstallGroup,
  TInstall,
  TInstallGroupRun,
} from '@/types'
import { hasLabelSelector, matchesSelector } from './label-selector'
import type { TLabelSelector } from './label-selector'

export type TInstallGroupMembership = 'default' | 'label_selector'

export type IDeploymentPlanInstall = Partial<TInstall> & { id: string }

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
  if (hasLabelSelector(group.label_selector)) return 'label_selector'
  return 'default'
}

const toPlanInstall = (
  install: Partial<TInstall> & { id: string }
): IDeploymentPlanInstall => ({
  ...install,
  id: install.id,
})

export const resolveDeploymentPlanStages = ({
  groups,
  installs,
  groupRuns,
}: {
  groups?: TAppBranchInstallGroup[] | null
  installs?: Array<Partial<TInstall>> | null
  groupRuns?: TInstallGroupRun[] | null
}): IDeploymentPlanStage[] => {
  const orderedGroups = [...(groups ?? [])]
    .map((group, index) => ({ group, index }))
    .sort(
      (a, b) => (a.group.order ?? 0) - (b.group.order ?? 0) || a.index - b.index
    )
    .map(({ group }) => group)

  const knownInstalls = (installs ?? []).filter(
    (
      install
    ): install is Partial<TInstall> & {
      id: string
    } => Boolean(install.id)
  )

  const runsByGroupId = new Map<string, TInstallGroupRun>()
  for (const run of groupRuns ?? []) {
    if (!run.install_group_id) continue
    runsByGroupId.set(run.install_group_id, run)
  }

  const membersByGroup = orderedGroups.map(() => [] as IDeploymentPlanInstall[])
  const defaultGroupIndex = orderedGroups.findIndex((group) => group.default)
  for (const install of knownInstalls) {
    let groupIndex = -1
    if (install.app_branch_group) {
      groupIndex = orderedGroups.findIndex(
        (group) => group.name === install.app_branch_group
      )
    } else {
      groupIndex = orderedGroups.findIndex(
        (group) =>
          hasLabelSelector(group.label_selector) &&
          matchesSelector(install.labels, group.label_selector)
      )
      if (groupIndex === -1) groupIndex = defaultGroupIndex
    }
    if (groupIndex !== -1) {
      membersByGroup[groupIndex].push(toPlanInstall(install))
    }
  }

  return orderedGroups.map((group, index) => {
    const membership = membershipOf(group)
    const members = membersByGroup[index]

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
      stage.autoApproveOnPoliciesPassing =
        group.auto_approve_on_policies_passing
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
