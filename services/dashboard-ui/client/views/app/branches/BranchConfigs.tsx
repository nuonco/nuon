import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getBranchConfigs } from '@/lib'
import { BranchProvider } from '@/providers/branch-provider'

const BranchConfigsContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const branchId = params.branchId as string

  const { data: configs, isLoading } = useQuery({
    queryKey: ['branch-configs', org.id, app.id, branchId],
    queryFn: () => getBranchConfigs({ orgId: org.id!, appId: app.id!, branchId }),
    enabled: !!org.id && !!app.id && !!branchId,
  })

  if (isLoading) {
    return <Text variant="body">Loading configs...</Text>
  }

  return (
    <div className="flex flex-col gap-6 p-6">
      <Text variant="h3" weight="strong">
        Branch Configs
      </Text>
      <Text variant="body">
        {configs?.length || 0} config{configs?.length !== 1 ? 's' : ''} found
      </Text>
    </div>
  )
}

export const BranchConfigs = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId}>
      <BranchConfigsContent />
    </BranchProvider>
  )
}
