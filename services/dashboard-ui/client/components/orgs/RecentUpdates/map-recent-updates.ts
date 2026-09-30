import type { TBranchPlanGroup } from '@/components/branches/BranchCards/BranchPlanDots'
import { githubCommitUrl } from '@/components/branches/shared/pr-link'
import type {
  TBranchActivityItem,
  TRunPendingApproval,
  TRunUpdatedInstall,
} from '@/components/orgs/BranchActivityFeed'
import { installHref } from '@/lib/install-path'
import type { TWorkflowStepApprovalType } from '@/types'
import { getApprovalType } from '@/utils/approval-utils'
import { getWorstStatusTheme } from '@/utils/status-utils'

export type TRecentConfigGroup = {
  id?: string
  name?: string
  order?: number
  hasSelector?: boolean
}

export type TRecentGroupRun = {
  install_group_id?: string
  install_group_name?: string
  total_installs?: number
  install_group?: {
    name?: string
    default?: boolean
    label_selector?: unknown
  }
  installs?: Array<{
    install_id?: string
    status?: string
    workflow_id?: string
  }>
}

export type TRecentApproval = {
  id?: string
  type?: TWorkflowStepApprovalType
  owner_id?: string
  owner_type?: string
  workflow_step?: {
    install_workflow_id?: string
    owner_id?: string
    owner_type?: string
  }
}

export type TRecentUpdateSource = {
  orgId: string
  nestedInstalls: boolean
  appId: string
  appName: string
  branchId: string
  branchName: string
  runId: string
  workflowId?: string
  status?: string
  createdAt?: string
  awaitingApproval?: boolean
  commitMessage?: string
  commitSha?: string
  commitAuthor?: string
  commitAvatarUrl?: string
  repo?: string
  configGroups?: TRecentConfigGroup[]
  groupRuns?: TRecentGroupRun[]
  installsById?: Record<string, { name?: string; health?: string }>
  approvals?: TRecentApproval[]
}

export type TRecentUpdatePayloadItem = Omit<
  TRecentUpdateSource,
  'orgId' | 'nestedInstalls'
>

export type TRecentUpdatesPayload = {
  updates: TRecentUpdatePayloadItem[]
}

const groupName = (groupRun: TRecentGroupRun, index: number) =>
  groupRun.install_group_name ||
  groupRun.install_group?.name ||
  `Group ${index + 1}`

const updatedInstalls = (
  source: TRecentUpdateSource,
  groupRuns: TRecentGroupRun[]
): TRunUpdatedInstall[] => {
  const installs: TRunUpdatedInstall[] = []
  groupRuns.forEach((groupRun, index) => {
    const name = groupName(groupRun, index)
    for (const row of groupRun.installs ?? []) {
      const id = row.install_id
      if (!id) continue
      const install = source.installsById?.[id]
      installs.push({
        id,
        name: install?.name || id,
        group: name,
        href: installHref({
          orgId: source.orgId,
          appId: source.appId,
          installId: id,
          nested: source.nestedInstalls,
        }),
        runStatus: row.status,
        health: install?.health,
        rolledOut: row.status === 'success',
      })
    }
  })
  return installs
}

const TERMINAL_RUN_STATUSES = new Set([
  'success',
  'failed',
  'error',
  'cancelled',
])

const groupStatus = (
  groupRun: TRecentGroupRun | undefined,
  runStatus: string | undefined
) => {
  const statuses = (groupRun?.installs ?? []).map((install) => install.status)
  if (statuses.length > 0) return getWorstStatusTheme(statuses).worstStatus
  if (groupRun) return TERMINAL_RUN_STATUSES.has(runStatus ?? '') ? 'noop' : 'pending'
  return TERMINAL_RUN_STATUSES.has(runStatus ?? '') ? 'not-attempted' : 'pending'
}

const planGroups = (
  configGroups: TRecentConfigGroup[],
  groupRuns: TRecentGroupRun[],
  installs: TRunUpdatedInstall[],
  runStatus: string | undefined
): TBranchPlanGroup[] => {
  const countByName = new Map<string, number>()
  for (const install of installs) {
    countByName.set(install.group, (countByName.get(install.group) ?? 0) + 1)
  }

  const ordered = [...configGroups].sort(
    (a, b) => (a.order ?? 0) - (b.order ?? 0)
  )
  if (ordered.length > 0) {
    return ordered.map((group, index) => {
      const name = group.name || `Group ${index + 1}`
      const groupRun = groupRuns.find(
        (run) =>
          (group.id && run.install_group_id === group.id) ||
          run.install_group_name === name
      )
      return {
        name,
        installs: countByName.get(name) || groupRun?.total_installs || 0,
        hasSelector: !!group.hasSelector,
        status: groupStatus(groupRun, runStatus),
      }
    })
  }

  return groupRuns.map((groupRun, index) => {
    const name = groupName(groupRun, index)
    return {
      name,
      installs:
        countByName.get(name) ||
        groupRun.total_installs ||
        groupRun.installs?.length ||
        0,
      hasSelector:
        !!groupRun.install_group?.label_selector ||
        !!groupRun.install_group?.default,
      status: groupStatus(groupRun, runStatus),
    }
  })
}

const approvalWorkflowIds = (groupRuns: TRecentGroupRun[]) => {
  const ids = new Set<string>()
  for (const groupRun of groupRuns) {
    for (const install of groupRun.installs ?? []) {
      if (install.workflow_id) ids.add(install.workflow_id)
    }
  }
  return ids
}

const belongsToRun = (
  approval: TRecentApproval,
  source: TRecentUpdateSource,
  childWorkflowIds: Set<string>
) => {
  if (approval.owner_id && approval.owner_id === source.runId) return true
  const stepWorkflowId = approval.workflow_step?.install_workflow_id
  if (!stepWorkflowId) return false
  if (source.workflowId && stepWorkflowId === source.workflowId) return true
  return childWorkflowIds.has(stepWorkflowId)
}

const approvalInstallId = (
  approval: TRecentApproval,
  groupRuns: TRecentGroupRun[]
) => {
  if (approval.owner_type === 'installs' && approval.owner_id) {
    return approval.owner_id
  }
  const step = approval.workflow_step
  if (step?.owner_type === 'installs' && step.owner_id) return step.owner_id
  const workflowId = step?.install_workflow_id
  if (!workflowId) return undefined
  for (const groupRun of groupRuns) {
    for (const install of groupRun.installs ?? []) {
      if (install.workflow_id === workflowId && install.install_id) {
        return install.install_id
      }
    }
  }
  return undefined
}

const pendingApprovals = (
  source: TRecentUpdateSource,
  groupRuns: TRecentGroupRun[]
): TRunPendingApproval[] => {
  const childWorkflowIds = approvalWorkflowIds(groupRuns)
  return (source.approvals ?? []).flatMap((approval) => {
    if (!approval.id || !belongsToRun(approval, source, childWorkflowIds)) {
      return []
    }
    const installId = approvalInstallId(approval, groupRuns)
    const install = installId ? source.installsById?.[installId] : undefined
    return [
      {
        id: approval.id,
        installName: install?.name || installId || source.branchName,
        type: approval.type ? getApprovalType(approval.type) : 'approval',
        href: installId
          ? installHref({
              orgId: source.orgId,
              appId: source.appId,
              installId,
              nested: source.nestedInstalls,
            })
          : undefined,
      },
    ]
  })
}

export const toRecentUpdateItem = (
  source: TRecentUpdateSource
): TBranchActivityItem | undefined => {
  if (!source.runId || !source.createdAt) return undefined

  const groupRuns = source.groupRuns ?? []
  const installs = updatedInstalls(source, groupRuns)
  const approvals = pendingApprovals(source, groupRuns)
  const appHref = `/${source.orgId}/apps/${source.appId}`
  const branchHref = `${appHref}/branches/${source.branchId}`
  const runHref = source.workflowId
    ? `${branchHref}/runs/${source.workflowId}`
    : undefined

  return {
    appId: source.appId,
    appName: source.appName,
    appHref,
    branchId: source.branchId,
    branchName: source.branchName,
    branchHref,
    runId: source.runId,
    runStatus:
      source.awaitingApproval || approvals.length > 0
        ? 'awaiting-approval'
        : source.status || 'pending',
    runCreatedAt: source.createdAt,
    runHref,
    commitMessage: source.commitMessage,
    commitSha: source.commitSha,
    commitAuthor: source.commitAuthor,
    commitAvatarUrl: source.commitAvatarUrl,
    commitHref: githubCommitUrl(source.repo, source.commitSha),
    planGroups: planGroups(
      source.configGroups ?? [],
      groupRuns,
      installs,
      source.status
    ),
    updatedInstalls: installs,
    pendingApprovals: approvals,
  }
}
