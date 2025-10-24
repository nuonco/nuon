import { WorkflowSteps as Steps } from '@/components/workflows/WorkflowSteps'
import { Empty } from '@/components/old/Empty'
import { getWorkflowById } from '@/lib'

// TODO(nnnat): this should fetch only the steps
// when we switch to the resigned workflow view
export const WorkflowSteps = async ({
  orgId,
  workflowId,
}: {
  orgId: string
  workflowId: string
}) => {
  const { data, error } = await getWorkflowById({
    orgId,
    workflowId,
  })

  return data && !error ? (
    <Steps initWorkflow={data} workflowId={workflowId} shouldPoll />
  ) : (
    <Empty
      emptyTitle="No workflow steps"
      emptyMessage="Waiting on workflow steps to generate"
      variant="history"
    />
  )
}
