import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'
import { AppBranchSwitcher } from './AppBranchSwitcher'

const LIMIT = 100

export const AppBranchSwitcherContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()

  const { data: result, isLoading } = useQuery({
    queryKey: ['app-branches', org.id, app.id, 'switcher', LIMIT],
    queryFn: () =>
      getAppBranches({
        orgId: org.id!,
        appId: app.id!,
        limit: LIMIT,
        offset: 0,
      }),
    enabled: !!org.id && !!app.id,
    placeholderData: keepPreviousData,
  })

  return (
    <AppBranchSwitcher
      branches={result?.data ?? []}
      currentBranch={branch}
      orgId={org.id!}
      appId={app.id!}
      isLoading={isLoading}
    />
  )
}
