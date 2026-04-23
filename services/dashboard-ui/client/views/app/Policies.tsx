import { useQuery } from '@tanstack/react-query'
import { PoliciesTable, policiesTableColumns } from '@/components/policies/PoliciesTable'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppPoliciesConfigs } from '@/lib'

export const Policies = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: policiesConfigs, isLoading } = useQuery({
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
      {isLoading ? (
        <TableSkeleton columns={policiesTableColumns} skeletonRows={5} />
      ) : (
        <PoliciesTable
          policies={policies}
          orgId={org?.id}
          appId={app?.id}
        />
      )}
    </div>
  )
}
