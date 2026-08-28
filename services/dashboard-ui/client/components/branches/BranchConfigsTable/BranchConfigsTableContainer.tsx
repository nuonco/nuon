import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getBranchConfigs } from '@/lib'
import { BranchConfigsTable } from './BranchConfigsTable'

export const BranchConfigsTableContainer = ({
  branchId,
}: {
  branchId: string
}) => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: configs, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-configs', org?.id, app?.id, branchId],
    queryFn: () =>
      getBranchConfigs({ orgId: org!.id, appId: app!.id, branchId }),
    enabled: !!org?.id && !!app?.id && !!branchId,
  })

  return (
    <BranchConfigsTable
      configs={configs ?? []}
      isLoading={isLoading}
      appId={app?.id}
    />
  )
}
