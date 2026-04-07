'use client'

import { useWorkflow } from '@/hooks/use-workflow'
import { StepBanner } from './step-details/StepBanner'
import { WorkflowHeaderContainer } from './workflow-details/WorkflowHeader'
import { WorkflowMetricsContainer } from './workflow-details/WorkflowMetrics'
import { WorkflowStatusSectionContainer } from './workflow-details/WorkflowStatusSection'
import { WorkflowDetailsSectionContainer } from './workflow-details/WorkflowDetailsSection'

export const WorkflowDetails = () => {
  const { failedSteps, workflow } = useWorkflow()

  return (
    <>
      <WorkflowHeaderContainer />

      <WorkflowMetricsContainer />

      <WorkflowStatusSectionContainer />

      <WorkflowDetailsSectionContainer />

      {failedSteps?.length > 0 &&
        failedSteps.map((failedStep) => (
          <div key={failedStep?.id} className="flex flex-col gap-4 mt-2">
            <StepBanner step={failedStep} planOnly />
          </div>
        ))}
    </>
  )
}
