import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import { AppSourceChip } from './AppSourceChip'

const toRepoHref = (repo?: string) => {
  if (!repo) return undefined
  return repo.startsWith('http') ? repo : `https://github.com/${repo}`
}

export const AppSourceChipContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: result, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branches-source', org?.id, app?.id],
    queryFn: () =>
      getAppBranches({
        orgId: org!.id!,
        appId: app!.id!,
        limit: 50,
        offset: 0,
      }),
    enabled: !!org?.id && !!app?.id,
  })

  const sources = (result?.data ?? []).flatMap((branch) => {
    const config = latestBranchConfig(branch)
    const vcs =
      config?.connected_github_vcs_config ?? config?.public_git_vcs_config
    return vcs?.repo ? [vcs] : []
  })
  const primary = sources[0]

  return (
    <AppSourceChip
      isLoading={isLoading}
      repo={primary?.repo}
      repoHref={toRepoHref(primary?.repo)}
    />
  )
}
