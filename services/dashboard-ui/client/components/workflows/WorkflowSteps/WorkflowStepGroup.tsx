import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { WorkflowStepRow } from './WorkflowStepRow'

export interface IWorkflowStepGroup {
  steps: TWorkflowStep[]
  approvalPrompt?: boolean
  planOnly?: boolean
}

export const WorkflowStepGroup = ({
  steps,
  approvalPrompt = false,
  planOnly = false,
}: IWorkflowStepGroup) => {
  const latest = steps.at(-1)
  const prior = steps
    .slice(0, -1)
    .map((step, idx) => ({ step, attemptNumber: idx + 1 }))
    .reverse()

  if (!latest) return null

  return (
    <Expand
      id={`step-group-${latest.id}`}
      interactiveHeading
      toggleLabel="Show previous attempts"
      className="border rounded-md"
      headerClassName="px-4 py-2"
      toggleContent={
        <Text variant="subtext" theme="neutral" nowrap>
          {prior.length} previous{' '}
          {prior.length === 1 ? 'attempt' : 'attempts'}
        </Text>
      }
      heading={
        <div className="flex-1 min-w-0">
          <WorkflowStepRow
            step={latest}
            approvalPrompt={approvalPrompt}
            planOnly={planOnly}
            showRetry
          />
        </div>
      }
    >
      <div className="flex flex-col divide-y border-t px-4">
        {prior.map(({ step, attemptNumber }) => (
          <div key={step.id} className="py-2 pl-8">
            <WorkflowStepRow
              step={step}
              approvalPrompt={approvalPrompt}
              planOnly={planOnly}
              showRetry={false}
              attemptNumber={attemptNumber}
              nested
            />
          </div>
        ))}
      </div>
    </Expand>
  )
}
