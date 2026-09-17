import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { useDismissedStepBanners } from '@/hooks/use-dismissed-step-banners'
import { useSurfaces } from '@/hooks/use-surfaces'
import { StepBanner } from '../step-details/StepBanner'
import {
  StepDetailPanel,
  getStepPanelSize,
  getStepPanelDetails,
} from '../step-details/StepDetailPanel'
import { WorkflowHeaderContainer } from '../workflow-details/WorkflowHeader'
import { WorkflowMetricsContainer } from '../workflow-details/WorkflowMetrics'
import { WorkflowStatusSectionContainer } from '../workflow-details/WorkflowStatusSection'
import { WorkflowDetailsSectionContainer } from '../workflow-details/WorkflowDetailsSection'
import type { TWorkflow, TWorkflowStep } from '@/types'

interface IWorkflowDetails {
  workflow: TWorkflow
  failedSteps: TWorkflowStep[]
}

export const WorkflowDetails = ({
  workflow,
  failedSteps,
}: IWorkflowDetails) => {
  const metadata = workflow?.status?.metadata
  const retriesExhausted = metadata?.retries_exhausted === true
  const stopped = metadata?.stopped === true
  const stopReason = metadata?.stop_reason as string | undefined
  const stepName = metadata?.step_name as string | undefined
  const failingStep = failedSteps?.find((step) => step?.name === stepName)
  const statusLine = workflow?.status?.status_human_description ?? ''
  const stopReasonShownInStatus = Boolean(
    stopReason && statusLine.includes(stopReason)
  )

  return (
    <div className="flex flex-col gap-2">
      {retriesExhausted && (
        <Banner theme="error">
          <div className="flex flex-col gap-1">
            <Text variant="body" weight="strong">
              Workflow cannot be retried
            </Text>
            <Text variant="subtext">
              This workflow has exhausted its retry limit
              {metadata?.max_retries
                ? ` (${metadata.max_retries} retries)`
                : ''}
              . Rerun the workflow to start fresh.
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
            {stopReasonShownInStatus ? null : (
              <Text variant="subtext">
                {stopReason ||
                  (metadata?.error_message as string) ||
                  'This workflow was stopped and cannot continue.'}
              </Text>
            )}
            {stepName ? (
              <StoppedStepLink stepName={stepName} step={failingStep} />
            ) : null}
          </div>
        </Banner>
      )}

      {failedSteps?.length > 0 && <FailedStepBanners steps={failedSteps} />}

      <WorkflowHeaderContainer />

      <WorkflowMetricsContainer />

      <WorkflowStatusSectionContainer />

      <WorkflowDetailsSectionContainer />
    </div>
  )
}

const stepDetailPanel = (step: TWorkflowStep) => (
  <StepDetailPanel
    panelKey={step.id}
    initStep={step}
    size={getStepPanelSize(step)}
    shouldPoll
    planOnly
  >
    {getStepPanelDetails(step)}
  </StepDetailPanel>
)

const StoppedStepLink = ({
  stepName,
  step,
}: {
  stepName: string
  step?: TWorkflowStep
}) => {
  const { addPanel } = useSurfaces()

  if (!step) {
    return <Text variant="subtext">Stopped at step {stepName}</Text>
  }

  return (
    <Text variant="subtext" flex>
      Stopped at step
      <Link
        isATag
        onClick={() => addPanel(stepDetailPanel(step), step.id)}
        className="cursor-pointer"
      >
        {stepName}
      </Link>
    </Text>
  )
}

const FailedStepBanners = ({ steps }: { steps: TWorkflowStep[] }) => {
  const [expanded, setExpanded] = useState(false)
  const { isDismissed, dismiss } = useDismissedStepBanners()
  const { addPanel } = useSurfaces()

  const visibleSteps = steps.filter((s) => !isDismissed(s.id))

  if (visibleSteps.length === 0) return null

  const isRetried = (step: TWorkflowStep) =>
    step?.status?.metadata?.auto_retried || step?.status?.metadata?.retried

  const openPanel = (step: TWorkflowStep) => {
    addPanel(stepDetailPanel(step), step.id)
  }

  if (visibleSteps.length === 1) {
    const step = visibleSteps[0]
    return (
      <div className="flex flex-col gap-4">
        <StepBanner
          step={step}
          planOnly
          onDismiss={isRetried(step) ? () => dismiss(step.id) : undefined}
          onViewDetails={() => openPanel(step)}
        />
      </div>
    )
  }

  const mostRecent = visibleSteps[visibleSteps.length - 1]
  const olderSteps = visibleSteps.slice(0, -1)

  const toggle = () => setExpanded((prev) => !prev)

  return (
    <div className="flex flex-col gap-2">
      <div
        role="button"
        tabIndex={0}
        aria-expanded={expanded}
        onClick={toggle}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            toggle()
          }
        }}
        className="flex items-center justify-between gap-3 cursor-pointer select-none focus:outline-none"
      >
        <Text
          as="span"
          variant="subtext"
          weight="strong"
          theme="error"
          flex
          nowrap
        >
          <Icon variant="WarningOctagonIcon" size={14} />
          {visibleSteps.length} steps failed
        </Text>
        <Text
          as="span"
          variant="subtext"
          weight="strong"
          flex
          nowrap
          className="shrink-0 text-primary-600 dark:text-primary-400"
        >
          {expanded ? 'Show less' : `Show ${olderSteps.length} more`}
          <Icon variant={expanded ? 'MinusIcon' : 'PlusIcon'} size={14} />
        </Text>
      </div>

      <div className="flex flex-col gap-4">
        <StepBanner
          step={mostRecent}
          planOnly
          onDismiss={
            isRetried(mostRecent) ? () => dismiss(mostRecent.id) : undefined
          }
          onViewDetails={() => openPanel(mostRecent)}
        />
      </div>

      {expanded &&
        olderSteps.map((step) => (
          <div key={step?.id} className="flex flex-col gap-4">
            <StepBanner
              step={step}
              planOnly
              onDismiss={isRetried(step) ? () => dismiss(step.id) : undefined}
              onViewDetails={() => openPanel(step)}
            />
          </div>
        ))}
    </div>
  )
}
