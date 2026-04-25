import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { StepBanner } from '../step-details/StepBanner'
import { WorkflowHeaderContainer } from '../workflow-details/WorkflowHeader'
import { WorkflowMetricsContainer } from '../workflow-details/WorkflowMetrics'
import { WorkflowStatusSectionContainer } from '../workflow-details/WorkflowStatusSection'
import { WorkflowDetailsSectionContainer } from '../workflow-details/WorkflowDetailsSection'
import type { TWorkflow, TWorkflowStep } from '@/types'

interface IWorkflowDetails {
  workflow: TWorkflow
  failedSteps: TWorkflowStep[]
}

export const WorkflowDetails = ({ workflow, failedSteps }: IWorkflowDetails) => {
  const metadata = workflow?.status?.metadata
  const retriesExhausted = metadata?.retries_exhausted === true
  const stopped = metadata?.stopped === true

  return (
    <>
      {retriesExhausted && (
        <Banner theme="error">
          <div className="flex flex-col gap-1">
            <Text variant="body" weight="strong">
              Workflow cannot be retried
            </Text>
            <Text variant="subtext">
              This workflow has exhausted its retry limit
              {metadata?.max_retries ? ` (${metadata.max_retries} retries)` : ''}.
              Rerun the workflow to start fresh.
            </Text>
          </div>
        </Banner>
      )}

      {stopped && !retriesExhausted && (
        <Banner theme="warn">
          <div className="flex flex-col gap-1">
            <Text variant="body" weight="strong">
              Workflow stopped
            </Text>
            <Text variant="subtext">
              {(metadata?.error_message as string) || 'This workflow was stopped and cannot continue.'}
            </Text>
          </div>
        </Banner>
      )}

      <WorkflowHeaderContainer />

      <WorkflowMetricsContainer />

      <WorkflowStatusSectionContainer />

      <WorkflowDetailsSectionContainer />

      {failedSteps?.length > 0 && (
        <FailedStepBanners steps={failedSteps} />
      )}
    </>
  )
}

const FailedStepBanners = ({ steps }: { steps: TWorkflowStep[] }) => {
  const [expanded, setExpanded] = useState(false)

  if (steps.length === 0) return null

  if (steps.length === 1) {
    return (
      <div className="flex flex-col gap-4 mt-2">
        <StepBanner step={steps[0]} planOnly />
      </div>
    )
  }

  const mostRecent = steps[steps.length - 1]
  const olderSteps = steps.slice(0, -1)

  return (
    <div className="flex flex-col gap-2 mt-2">
      <div className="flex flex-col gap-4">
        <StepBanner step={mostRecent} planOnly />
      </div>

      {olderSteps.length > 0 && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setExpanded(!expanded)}
          className="self-start"
        >
          {expanded
            ? 'Hide older errors'
            : `${olderSteps.length} more error${olderSteps.length > 1 ? 's' : ''}`}
        </Button>
      )}

      {expanded &&
        olderSteps.map((step) => (
          <div key={step?.id} className="flex flex-col gap-4">
            <StepBanner step={step} planOnly />
          </div>
        ))}
    </div>
  )
}
