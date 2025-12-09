'use client'

import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { Duration } from '@/components/common/Duration'
import { useWorkflow } from '@/hooks/use-workflow'

export const WorkflowMetrics = () => {
  const { 
    workflow,
    pendingApprovalsCount, 
    discardedStepsCount, 
    completedStepsCount, 
    totalSteps 
  } = useWorkflow()

  const showPendingApprovals = 
    workflow.approval_option === 'prompt' && 
    !workflow?.plan_only

  return (
    <div className="flex flex-col md:flex-row gap-2 md:items-center justify-between">
      <div className="flex flex-wrap md:items-center gap-2 md:gap-6">
        <LabeledValue label="Elapsed time">
          <Duration nanoseconds={workflow?.execution_time} variant="base" />
        </LabeledValue>

        {workflow.plan_only && (
          <LabeledValue label="Mode">
            <Tooltip
              position="right"
              showIcon
              tipContent={
                <span className="flex flex-col w-66">
                  <Text weight="strong">Drift scan</Text>
                  <Text variant="subtext" className="text-nowrap">
                    Generate the workflow script without executing to detect
                    any drift between the app configuration and this install.
                  </Text>
                </span>
              }
            >
              <Text variant="base">Drift scan</Text>
            </Tooltip>
          </LabeledValue>
        )}
      </div>

      <div className="flex flex-wrap md:items-center gap-2 md:gap-6">
        {showPendingApprovals && (
          <LabeledValue label="Pending approvals">
            <Text variant="base">{pendingApprovalsCount}</Text>
          </LabeledValue>
        )}

        <LabeledValue label="Discarded">
          <Text variant="base">{discardedStepsCount}</Text>
        </LabeledValue>

        <LabeledValue label="Completed">
          <Text variant="base">{completedStepsCount}</Text>
        </LabeledValue>

        <LabeledValue label="Total steps">
          <Text variant="base">{totalSteps}</Text>
        </LabeledValue>
      </div>
    </div>
  )
}