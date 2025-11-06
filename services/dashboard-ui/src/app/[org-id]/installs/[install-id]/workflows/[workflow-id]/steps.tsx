import { Text } from '@/components/common/Text'
import { WorkflowSteps as Steps } from '@/components/workflows/WorkflowSteps'
import { getWorkflowSteps } from '@/lib'

interface ILoadWorkflowSteps {
  offset: string
  orgId: string
  workflowId: string
}

export async function WorkflowSteps({
  orgId,
  offset,
  workflowId,
}: ILoadWorkflowSteps) {
  const {
    data: steps,
    error,
    headers,
  } = await getWorkflowSteps({ orgId, workflowId, offset })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return steps && !error ? (
    <>
      <Steps initWorkflowSteps={steps} shouldPoll workflowId={workflowId} />
    </>
  ) : (
    <WorkflowStepsError />
  )
}

export const WorkflowStepsError = () => (
  <div className="w-full">
    <Text>Error fetching recenty workflows activity </Text>
  </div>
)
