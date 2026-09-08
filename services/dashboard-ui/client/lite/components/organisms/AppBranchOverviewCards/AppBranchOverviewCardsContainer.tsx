import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getAppInstalls } from '@/lib'
import { useApp } from '../../../providers/app-provider'
import { useAppBranch } from '../../../providers/app-branch-provider'
import { useOrg } from '../../../providers/org-provider'
import { AppBranchOverviewCards } from './AppBranchOverviewCards'

const INSTALL_LIMIT = 100

export const AppBranchOverviewCardsContainer = () => {
  const { orgId } = useOrg()
  const { appId } = useApp()
  const { branch, branchId, isLoading: isLoadingBranch } = useAppBranch()
  const { data: result, isLoading: isLoadingInstalls } = useQuery({
    queryKey: ['app-installs', orgId, appId, branchId, 'overview'],
    queryFn: () =>
      getAppInstalls({
        orgId: orgId!,
        appId: appId!,
        app_branch_id: branchId!,
        limit: INSTALL_LIMIT,
        offset: 0,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    placeholderData: keepPreviousData,
    refetchInterval: 20_000,
  })

  return (
    <AppBranchOverviewCards
      branch={branch}
      installCount={result?.data?.length}
      hasMoreInstalls={result?.pagination?.hasNext}
      isLoading={isLoadingBranch || isLoadingInstalls}
    />
  )
}
