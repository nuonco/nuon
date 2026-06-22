import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { WorkflowStepRow } from './WorkflowStepRow'

export interface IWorkflowStepRoundGroup {
  steps: TWorkflowStep[]
  approvalPrompt?: boolean
  planOnly?: boolean
}

const roundOf = (step: TWorkflowStep) => step?.group_retry_idx ?? 0

export const WorkflowStepRoundGroup = ({
  steps,
  approvalPrompt = false,
  planOnly = false,
}: IWorkflowStepRoundGroup) => {
  const roundsByIdx = new Map<number, TWorkflowStep[]>()
  for (const step of steps) {
    const idx = roundOf(step)
    const round = roundsByIdx.get(idx)
    if (round) {
      round.push(step)
    } else {
      roundsByIdx.set(idx, [step])
    }
  }

  const rounds = [...roundsByIdx.entries()]
    .map(([idx, roundSteps]) => ({ idx, steps: roundSteps }))
    .sort((a, b) => a.idx - b.idx)

  const current = rounds.at(-1)
  const prior = rounds.slice(0, -1).reverse()

  if (!current) return null

  return (
    <Expand
      id={`step-round-group-${current.steps[0]?.id}`}
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
        <div className="flex-1 min-w-0 flex flex-col gap-3">
          {current.steps.map((step) => (
            <WorkflowStepRow
              key={step.id}
              step={step}
              approvalPrompt={approvalPrompt}
              planOnly={planOnly}
              showRetry
            />
          ))}
        </div>
      }
    >
      <div className="flex flex-col divide-y border-t">
        {prior.map((round) => (
          <div key={round.idx} className="flex flex-col gap-3 px-4 py-4 pl-8">
            <Text variant="label" theme="neutral" weight="strong">
              Attempt {round.idx + 1}
            </Text>
            {round.steps.map((step) => (
              <WorkflowStepRow
                key={step.id}
                step={step}
                approvalPrompt={approvalPrompt}
                planOnly={planOnly}
                showRetry={false}
                nested
              />
            ))}
          </div>
        ))}
      </div>
    </Expand>
  )
}
