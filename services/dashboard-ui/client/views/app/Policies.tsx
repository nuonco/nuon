import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { PoliciesTable, policiesTableColumns } from '@/components/policies/PoliciesTable'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useBranchScopedAppConfig } from '@/hooks/use-branch-scoped-app-config'
import { useOrg } from '@/hooks/use-org'
import { getAppPoliciesConfigs } from '@/lib'
import type { TAppPolicyConfig } from '@/types'

export const Policies = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branchId } = useParams()

  const { data: policiesConfigs, isLoading: isLoadingPolicies } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-policies-configs', org?.id, app?.id],
    queryFn: () => getAppPoliciesConfigs({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id && !branchId,
  })

  const { appConfig, isLoading: isLoadingBranchConfig } =
    useBranchScopedAppConfig({
      orgId: org?.id,
      appId: app?.id,
      branchId,
    })

  const latestConfig = policiesConfigs
    ?.slice()
    .sort((a, b) => {
      const dateA = a.created_at ? new Date(a.created_at).getTime() : 0
      const dateB = b.created_at ? new Date(b.created_at).getTime() : 0
      return dateB - dateA
    })
    .at(0)
  const policies = (
    branchId ? (appConfig?.policies?.policies ?? []) : (latestConfig?.policies ?? [])
  ) as TAppPolicyConfig[]
  const isLoading = branchId ? isLoadingBranchConfig : isLoadingPolicies

  return (
    <div className="flex flex-auto">
      <PageTitle segments={['Policies', app?.name]} />
      {isLoading ? (
        <TableSkeleton columns={policiesTableColumns} skeletonRows={5} />
      ) : (
        <PoliciesTable
          policies={policies}
          orgId={org?.id}
          appId={app?.id}
          branchId={branchId}
        />
      )}
    </div>
  )
}
