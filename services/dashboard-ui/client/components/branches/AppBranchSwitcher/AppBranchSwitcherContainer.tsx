import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useLocation } from 'react-router'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'
import { branchSwitchSectionPath } from '@/utils/branch-utils'
import { AppBranchSwitcher } from './AppBranchSwitcher'

const LIMIT = 100

export const AppBranchSwitcherContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const { pathname } = useLocation()

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
      sectionPath={
        branch?.id
          ? branchSwitchSectionPath(
              pathname,
              `/${org.id}/apps/${app.id}/branches/${branch.id}`
            )
          : ''
      }
      isLoading={isLoading}
    />
  )
}
