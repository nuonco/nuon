import { Text } from '@/components/common/Text'
import { WorkflowSteps as Steps } from '@/components/workflows/WorkflowSteps'
import { getWorkflowSteps } from '@/lib'

interface ILoadWorkflowSteps {
  approvalPrompt?: boolean
  offset: string
  orgId: string
  planOnly?: boolean
  workflowId: string
}

export async function WorkflowSteps({
  approvalPrompt = false,
  orgId,
  offset,
  planOnly = false,
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
      <Steps
        approvalPrompt={approvalPrompt}
        initWorkflowSteps={steps}
        planOnly={planOnly}
        shouldPoll
        workflowId={workflowId}
      />
    </>
  ) : (
    <WorkflowStepsError />
  )
}

export const WorkflowStepsError = () => (
  <div className="w-full">
    <Text>Error fetching recent workflows activity </Text>
  </div>
)
