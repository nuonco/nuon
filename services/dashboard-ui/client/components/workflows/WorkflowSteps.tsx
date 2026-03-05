import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { PolicyCountsBadge } from '@/components/workflows/step-details/PolicyCountsBadge'
import { StepButtons } from '@/components/workflows/step-details/StepButtons'
import { StepDetailPanelButton } from '@/components/workflows/step-details/StepDetailPanel'
import { StepTitle } from '@/components/workflows/step-details/StepTitle'
import { useOrg } from '@/hooks/use-org'
import { useWorkflow } from '@/hooks/use-workflow'
import { getWorkflowSteps } from '@/lib'
import type { TWorkflowStep } from '@/types'
import { getStepBadge } from '@/utils/workflow-utils'

interface IWorkflowSteps {
  approvalPrompt?: boolean
  planOnly?: boolean
  pollInterval?: number
  shouldPoll?: boolean
  workflowId: string
}

export const WorkflowSteps = ({
  approvalPrompt = false,
  planOnly = false,
  pollInterval = 4000,
  shouldPoll = false,
  workflowId,
}: IWorkflowSteps) => {
  const { org } = useOrg()
  const { workflow } = useWorkflow()
  const [searchName, setSearchName] = useState<string>('')

  const shouldStopPolling = workflow?.finished || workflow?.status?.status === 'cancelled'
  const effectiveShouldPoll = shouldPoll && !shouldStopPolling

  const { data: workflowSteps = [] } = useQuery<TWorkflowStep[]>({
    queryKey: ['workflow-steps', org?.id, workflowId],
    queryFn: () => getWorkflowSteps({ orgId: org.id, workflowId }),
    refetchOnMount: 'always',
    refetchInterval: effectiveShouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!workflowId,
  })

  const filteredSteps = workflowSteps
    .filter((step) => step.execution_type !== 'hidden')
    .filter((step) => step.name.includes(searchName))

  return (
    <div className="flex flex-col gap-6">
      <SearchInput
        placeholder="Search workflow steps"
        value={searchName}
        onChange={setSearchName}
      />
      <div className="flex flex-col gap-4">
        {filteredSteps.length ? (
          filteredSteps.map((step) => {
            const badgeConfig = getStepBadge(step, approvalPrompt)

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
                  step.status.status === 'pending' ? null : (
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

                <StepButtons isApproveAll={!approvalPrompt} step={step} />
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
