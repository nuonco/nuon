import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useWorkflow } from '@/hooks/use-workflow'
import { getWorkflowQueuePosition } from '@/lib'
import { WorkflowStatusSection } from './WorkflowStatusSection'

export const WorkflowStatusSectionContainer = () => {
  const { workflow } = useWorkflow()
  const { org } = useOrg()

  const isPending =
    workflow?.status?.status === 'pending' ||
    workflow?.status?.status === 'queued'

  const { data: queuePosition } = useQuery({
    queryKey: ['workflow-queue-position', org?.id, workflow?.id],
    queryFn: () =>
      getWorkflowQueuePosition({
        workflowId: workflow!.id,
        orgId: org!.id,
      }),
    enabled: !!org?.id && !!workflow?.id && isPending,
    refetchInterval: isPending ? 5000 : false,
  })

  if (!workflow) return null

  return (
    <WorkflowStatusSection workflow={workflow} queuePosition={queuePosition} />
  )
}
