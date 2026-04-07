import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useWorkflow } from '@/hooks/use-workflow'
import { getWorkflowSteps } from '@/lib'
import type { TWorkflowStep } from '@/types'
import { WorkflowSteps, WorkflowStepsSkeleton } from './WorkflowSteps'

export { WorkflowStepsSkeleton }

interface IWorkflowStepsContainer {
  approvalPrompt?: boolean
  planOnly?: boolean
  pollInterval?: number
  shouldPoll?: boolean
  workflowId: string
}

export const WorkflowStepsContainer = ({
  approvalPrompt = false,
  planOnly = false,
  pollInterval = 4000,
  shouldPoll = false,
  workflowId,
}: IWorkflowStepsContainer) => {
  const { org } = useOrg()
  const { workflow } = useWorkflow()

  const shouldStopPolling = workflow?.finished || workflow?.status?.status === 'cancelled'
  const effectiveShouldPoll = shouldPoll && !shouldStopPolling

  const { data: workflowSteps = [] } = useQuery<TWorkflowStep[]>({
    queryKey: ['workflow-steps', org?.id, workflowId],
    queryFn: () => getWorkflowSteps({ orgId: org.id, workflowId }),
    refetchOnMount: 'always',
    refetchInterval: effectiveShouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!workflowId,
  })

  return (
    <WorkflowSteps
      approvalPrompt={approvalPrompt}
      planOnly={planOnly}
      workflowSteps={workflowSteps}
    />
  )
}
