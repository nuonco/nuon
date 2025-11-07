'use client'

import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { StepButtons } from '@/components/workflows/step-details/StepButtons'
import { StepDetailPanelButton } from '@/components/workflows/step-details/StepDetailPanel'
import { StepTitle } from '@/components/workflows/step-details/StepTitle'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TWorkflowStep } from '@/types'
import { getStepBadge } from '@/utils/workflow-utils'

interface IWorkflowSteps extends IPollingProps {
  initWorkflowSteps: TWorkflowStep[]
  planOnly?: boolean
  workflowId: string
}

export const WorkflowSteps = ({
  initWorkflowSteps,
  planOnly = false,
  pollInterval = 4000,
  shouldPoll = false,
  workflowId,
}: IWorkflowSteps) => {
  const { org } = useOrg()
  const [searchName, setSearchName] = useState<string>('')
  const { data: workflowSteps } = usePolling<TWorkflowStep[]>({
    path: `/api/orgs/${org?.id}/workflows/${workflowId}/steps`,
    shouldPoll,
    initData: initWorkflowSteps,
    pollInterval,
  })

  const filteredSteps = workflowSteps
    ? workflowSteps
        .filter((step) => step.execution_type !== 'hidden')
        .filter((step) => step.name.includes(searchName))
        .sort((a, b) => a.idx - b.idx)
    : []

  return (
    <div className="flex flex-col gap-6">
      <SearchInput
        placeholder="Search workflow steps"
        value={searchName}
        onChange={setSearchName}
      />
      <div className="flex flex-col gap-4">
        {filteredSteps && filteredSteps?.length ? (
          filteredSteps.map((step) => {
            const badgeConfig = getStepBadge(step)

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

                  {(step.execution_type === 'system' &&
                    !step.step_target_type) ||
                  step.status.status === 'pending' ? null : (
                    <StepDetailPanelButton step={step} planOnly={planOnly} />
                  )}

                  {step?.finished ? (
                    <Text variant="subtext" theme="neutral">
                      Completed in{' '}
                      <Duration
                        variant="subtext"
                        nanoseconds={step?.execution_time}
                      />
                    </Text>
                  ) : null}
                </div>

                <StepButtons step={step} />
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
            emptyTitle={
              workflowSteps.length
                ? 'No steps found'
                : 'Workflow steps not available'
            }
          />
        )}
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
