import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { PolicyCountsBadge } from '@/components/workflows/step-details/PolicyCountsBadge'
import { StepButtons } from '@/components/workflows/step-details/StepButtons'
import { StepDetailPanelButton } from '@/components/workflows/step-details/StepDetailPanel'
import { StepTitle } from '@/components/workflows/step-details/StepTitle'
import type { TWorkflowStep } from '@/types'
import { getStepBadge, getStepKind } from '@/utils/workflow-utils'

export interface IWorkflowSteps {
  approvalPrompt?: boolean
  planOnly?: boolean
  workflowSteps: TWorkflowStep[]
  eagerStepsLoaded?: boolean
  allStepsLoaded?: boolean
}

export const WorkflowSteps = ({
  approvalPrompt = false,
  planOnly = false,
  workflowSteps,
  eagerStepsLoaded = false,
  allStepsLoaded = false,
}: IWorkflowSteps) => {
  const [searchName, setSearchName] = useState<string>('')

  const filteredSteps = workflowSteps
    .filter((step) => step.execution_type !== 'hidden')
    .filter((step) => step.name.includes(searchName))

  // Retrying a step appends a new attempt row that shares the same logical
  // "kind" (group + name) as the prior attempts. Only the latest attempt of a
  // kind should expose retry/skip controls, so track the last index per kind.
  const lastIdxByKind = new Map<string, number>()
  filteredSteps.forEach((step, idx) => {
    lastIdxByKind.set(getStepKind(step), idx)
  })

  return (
    <div className="flex flex-col gap-6">
      <SearchInput
        placeholder="Search workflow steps"
        value={searchName}
        onChange={setSearchName}
      />
      <div className="flex flex-col gap-4">
        {filteredSteps.length ? (
          filteredSteps.map((step, idx) => {
            const badgeConfig = getStepBadge(step, approvalPrompt, planOnly)
            const isLatestAttempt = lastIdxByKind.get(getStepKind(step)) === idx

            return (
              <div
                key={step.id}
                className="flex flex-col md:flex-row md:items-center gap-4 border px-4 py-2 rounded-md"
              >
                <StepTitle step={step} />

                <div className="flex items-center flex-wrap gap-2 md:gap-4">
                  {badgeConfig?.children ? (
                    <Badge {...badgeConfig} size="sm" />
                  ) : null}

                  <PolicyCountsBadge step={step} />

                  {(step.execution_type === 'system' && !step.step_target_type) ||
                  step.status.status === 'pending' ||
                  step.status.status === 'not-attempted' ? null : (
                    <StepDetailPanelButton
                      approvalPrompt={approvalPrompt}
                      step={step}
                      planOnly={planOnly}
                    />
                  )}

                  {step?.finished ? (
                    <Text variant="subtext" theme="neutral">
                      Completed in{' '}
                      <Duration variant="subtext" nanoseconds={step?.execution_time} />
                    </Text>
                  ) : null}
                </div>

                <StepButtons
                  isApproveAll={!approvalPrompt}
                  showRetry={isLatestAttempt}
                  step={step}
                />
              </div>
            )
          })
        ) : (
          <EmptyState
            variant="table"
            emptyMessage={
              workflowSteps.length
                ? 'No workflow steps match your search. Try adjusting your search criteria.'
                : 'Steps will appear here once the workflow has been generated.'
            }
            emptyTitle={workflowSteps.length ? 'No steps found' : 'Workflow steps not available'}
          />
        )}

        {eagerStepsLoaded && !allStepsLoaded ? (
          <div className="flex flex-col gap-4">
            {Array.from({ length: 2 }).map((_, idx) => (
              <Skeleton key={idx} height="44px" width="100%" className="rounded-md" />
            ))}
            <div className="flex items-center justify-center gap-2 py-1">
              <Loading />
              <Text variant="body" theme="neutral">
                More steps generating…
              </Text>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  )
}

export const WorkflowStepsSkeleton = () => {
  return (
    <div className="flex flex-col gap-4">
      {Array.from({ length: 8 }).map((_, idx) => (
        <Skeleton key={idx} height="44px" width="100%" />
      ))}
    </div>
  )
}
