import { useSearchParams } from 'react-router'
import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getAppBranches,
  getAppInstalls,
  getBranchInstallGroupRuns,
} from '@/lib'
import type { TAppBranch, TInstall, TInstallGroupRun } from '@/types'
import { BranchManagementDropdown } from '@/components/branches/management/BranchManagementDropdown'
import { CreateBranchButton } from '@/components/branches/CreateBranchModal'
import { latestBranchConfig } from '@/utils/branch-utils'
import { BranchCards } from './BranchCards'
import type { TBranchCardData } from './BranchCard'
import type { TBranchRolloutInstall } from './BranchRolloutStatus'

const LIMIT = 20
const INSTALL_LIMIT = 100
const TERMINAL_RUN_STATUSES = new Set([
  'success',
  'failed',
  'error',
  'cancelled',
])

export const countInstallsByBranch = (installs: TInstall[]) =>
  installs.reduce<Record<string, number>>((counts, install) => {
    const branchId = install.app_branch_id ?? install.app_branch?.id
    if (branchId) counts[branchId] = (counts[branchId] ?? 0) + 1
    return counts
  }, {})

export const rolloutInstallsFromGroupRuns = (
  groupRuns: TInstallGroupRun[],
  installNames: Record<string, string | undefined>
): TBranchRolloutInstall[] =>
  groupRuns.flatMap((groupRun) =>
    (groupRun.installs ?? []).flatMap((install) =>
      install.install_id
        ? [
            {
              id: install.install_id,
              name: installNames[install.install_id] ?? install.install_id,
              status: install.status || 'pending',
            },
          ]
        : []
    )
  )

export function parseBranchToCardData(
  branch: TAppBranch,
  orgId: string,
  appId: string,
  installCounts?: { counts: Record<string, number>; isPartial: boolean },
  rolloutInstalls?: TBranchRolloutInstall[]
): TBranchCardData {
  const config = latestBranchConfig(branch)
  const vcs =
    config?.connected_github_vcs_config ?? config?.public_git_vcs_config
  const installGroups = config?.install_groups ?? []
  const href = `/${orgId}/apps/${appId}/branches/${branch.id}`

  return {
    branchId: branch.id || '',
    name: branch.name || '',
    href,
    repo: vcs?.repo,
    repoBranch: vcs?.branch,
    latestRun: branch.latest_run
      ? {
          href: branch.latest_run.workflow_id
            ? `${href}/runs/${branch.latest_run.workflow_id}`
            : undefined,
          status: branch.latest_run.status || 'pending',
          commitMessage: branch.latest_run.vcs_connection_commit?.message,
          author: branch.latest_run.vcs_connection_commit?.author_name,
          avatarUrl: branch.latest_run.vcs_connection_commit?.author_avatar_url,
          sha: branch.latest_run.vcs_connection_commit?.sha,
          createdAt: branch.latest_run.created_at,
          awaitingApproval: branch.latest_run.awaiting_approval,
        }
      : undefined,
    planGroups: [...installGroups]
      .sort((a, b) => (a.order ?? 0) - (b.order ?? 0))
      .map((group, index) => ({
        name: group.name || `Group ${index + 1}`,
        installs: 0,
        hasSelector: !!group.label_selector || !!group.default,
      })),
    installCount: installCounts
      ? (installCounts.counts[branch.id ?? ''] ?? 0)
      : undefined,
    installCountIsPartial: installCounts?.isPartial,
    rolloutInstalls,
    action: (
      <BranchManagementDropdown branch={branch} appId={appId} orgId={orgId} />
    ),
  }
}

export const BranchCardsContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const { org } = useOrg()
  const { app } = useApp()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['app-branches', org.id, app.id, offset],
    queryFn: () =>
      getAppBranches({ orgId: org.id!, appId: app.id!, limit: LIMIT, offset }),
    enabled: !!org.id && !!app.id,
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const { data: installsResult } = useQuery({
    queryKey: ['app-installs', org.id, app.id],
    queryFn: () =>
      getAppInstalls({ orgId: org.id!, appId: app.id!, limit: INSTALL_LIMIT }),
    enabled: !!org.id && !!app.id,
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })
  const installs = installsResult?.data ?? []
  const installCounts = installsResult
    ? {
        counts: countInstallsByBranch(installs),
        isPartial: installsResult.pagination.hasNext,
      }
    : undefined
  const installNames = Object.fromEntries(
    installs.map((install) => [install.id, install.name])
  )

  const branches = result?.data ?? []
  const groupRunResults = useQueries({
    queries: branches.map((branch) => {
      const runId = branch.latest_run?.id
      const isTerminal = TERMINAL_RUN_STATUSES.has(
        branch.latest_run?.status ?? ''
      )
      return {
        queryKey: [
          'branch-install-group-runs',
          org.id,
          app.id,
          branch.id,
          runId,
        ],
        queryFn: () =>
          getBranchInstallGroupRuns({
            orgId: org.id!,
            appId: app.id!,
            branchId: branch.id!,
            runId: runId!,
          }),
        enabled: !!org.id && !!app.id && !!branch.id && !!runId,
        placeholderData: keepPreviousData,
        refetchInterval: shouldPoll && !isTerminal ? pollInterval : false,
      }
    }),
  })

  return (
    <BranchCards
      cards={branches.map((branch, index) => {
        const groupRuns = groupRunResults[index]?.data
        return parseBranchToCardData(
          branch,
          org.id!,
          app.id!,
          installCounts,
          groupRuns
            ? rolloutInstallsFromGroupRuns(groupRuns, installNames)
            : undefined
        )
      })}
      isLoading={isLoading}
      emptyAction={<CreateBranchButton />}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
