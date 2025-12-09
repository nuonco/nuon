'use client'

import { useWorkflow } from '@/hooks/use-workflow'
import { StepBanner } from './step-details/StepBanner'
import { WorkflowHeader } from './workflow-details/WorkflowHeader'
import { WorkflowMetrics } from './workflow-details/WorkflowMetrics'
import { WorkflowStatusSection } from './workflow-details/WorkflowStatusSection'
import { WorkflowDetailsSection } from './workflow-details/WorkflowDetailsSection'

export const WorkflowDetails = () => {
  const { failedSteps } = useWorkflow()

  return (
    <>
      <WorkflowHeader />

      <WorkflowMetrics />

      <WorkflowStatusSection />

      <WorkflowDetailsSection />

      {failedSteps?.length > 0 && (
        <>
          {failedSteps.map((failedStep) => (
            <div key={failedStep?.id} className="mt-2">
              <StepBanner step={failedStep} planOnly />
            </div>
          ))}
        </>
      )}
    </>
  )
}