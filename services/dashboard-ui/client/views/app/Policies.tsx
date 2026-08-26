import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { PoliciesTable, policiesTableColumns } from '@/components/policies/PoliciesTable'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppPoliciesConfigs } from '@/lib'

export const Policies = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branchId } = useParams()

  const { data: policiesConfigs, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-policies-configs', org?.id, app?.id],
    queryFn: () => getAppPoliciesConfigs({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const latestConfig = policiesConfigs
    ?.slice()
    .sort((a, b) => {
      const dateA = a.created_at ? new Date(a.created_at).getTime() : 0
      const dateB = b.created_at ? new Date(b.created_at).getTime() : 0
      return dateB - dateA
    })
    .at(0)
  const policies = latestConfig?.policies ?? []

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
