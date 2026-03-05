import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getBranchWorkflowRuns } from '@/lib'
import { BranchProvider } from '@/providers/branch-provider'

const BranchRunsContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const branchId = params.branchId as string

  const { data: runs, isLoading } = useQuery({
    queryKey: ['branch-runs', org.id, app.id, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({ orgId: org.id!, appId: app.id!, branchId }),
    enabled: !!org.id && !!app.id && !!branchId,
  })

  if (isLoading) {
    return <Text variant="body">Loading runs...</Text>
  }

  return (
    <div className="flex flex-col gap-6 p-6">
      <Text variant="h3" weight="strong">
        Workflow Runs
      </Text>
      <Text variant="body">
        {runs?.length || 0} run{runs?.length !== 1 ? 's' : ''} found
      </Text>
    </div>
  )
}

export const BranchRuns = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId}>
      <BranchRunsContent />
    </BranchProvider>
  )
}
