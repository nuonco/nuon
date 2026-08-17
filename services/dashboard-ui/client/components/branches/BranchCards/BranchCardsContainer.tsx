import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'
import type { TAppBranch } from '@/types'
import { BranchManagementDropdown } from '@/components/branches/management/BranchManagementDropdown'
import { CreateBranchButton } from '@/components/branches/CreateBranchModal'
import { latestBranchConfig } from '@/utils/branch-utils'
import { BranchCards } from './BranchCards'
import type { TBranchCardData } from './BranchCard'

const LIMIT = 20

export function parseBranchToCardData(
  branch: TAppBranch,
  orgId: string,
  appId: string
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
    managedBy: branch.managed_by,
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
    planSummary: {
      groups: installGroups.length,
      installs: installGroups.reduce(
        (sum, group) => sum + (group.install_ids?.length ?? 0),
        0
      ),
      hasSelector: installGroups.some((group) => !!group.label_selector),
    },
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

  return (
    <BranchCards
      cards={(result?.data ?? []).map((branch) =>
        parseBranchToCardData(branch, org.id!, app.id!)
      )}
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
