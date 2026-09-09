import { Navigate } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getAppBranches } from '@/lib'
import { Spinner } from '../components/atoms/Spinner'
import { Text } from '../components/atoms/Text'
import { getLastAppBranch } from '../utils/app-branch-session'
import { resolveAppHref } from '../utils/hrefs'
import { useApp } from '../providers/app-provider'
import { useOrg } from '../providers/org-provider'

export const AppResolver = () => {
  const { orgId } = useOrg()
  const { appId } = useApp()
  const { data: result, isLoading, error } = useQuery({
    queryKey: ['app-branches', orgId, appId],
    queryFn: () =>
      getAppBranches({
        orgId: orgId!,
        appId: appId!,
        limit: 50,
        offset: 0,
      }),
    enabled: !!orgId && !!appId,
  })

  if (!orgId || !appId || isLoading) {
    return (
      <div className="flex min-h-40 items-center justify-center">
        <Spinner size={20} label="Loading app" />
      </div>
    )
  }

  if (error) {
    return (
      <Text variant="caption" color="secondary">
        App branches failed to load
      </Text>
    )
  }

  const branchIds = (result?.data ?? [])
    .map((branch) => branch?.id)
    .filter((id): id is string => !!id)

  return (
    <Navigate
      replace
      to={resolveAppHref({
        orgId,
        appId,
        branchIds,
        lastBranchId: getLastAppBranch(orgId, appId),
      })}
    />
  )
}
