import { useState, type ReactNode } from 'react'
import { APPROVAL_BANNER_COPY } from '@/components/approvals/ApprovalBanner'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { Tooltip } from '@/components/common/Tooltip'
import { RunningClock } from '@/components/workflows/RunningClock'
import { LogStreamContext } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import type { TOTELLog, TWorkflowStep } from '@/types'
import { cn } from '@/utils/classnames'
import { StepTooltipContent } from './StepTooltip'
import { getWorkflowStepTitle } from '@/utils/workflow-utils'

const CURRENT_STATUSES = new Set([
  'in-progress',
  'approval-awaiting',
  'approval-denied',
  'error',
])

const PLAN_PANEL_ID = 'run-plan-panel'

const LIVE_STATUSES = new Set(['in-progress', 'approval-awaiting'])

const NOT_STARTED_STATUSES = new Set(['pending', 'not-attempted'])

const EMPTY_COPY: Record<string, { title: string; message: string }> = {
  pending: {
    title: 'Waiting to start',
    message: 'Logs appear once an earlier step finishes.',
  },
  'not-attempted': {
    title: 'Not attempted',
    message: 'This step did not run.',
  },
}

export interface IRunView {
  steps: TWorkflowStep[]
  logs: Record<string, TOTELLog[]>
  onApprove?: () => void
  onDeny?: () => void
  renderPlan?: (step: TWorkflowStep) => ReactNode
}

const StepTiming = ({ step }: { step: TWorkflowStep }) => {
  if (step.finished)
    return (
      <Duration
        variant="subtext"
        theme="neutral"
        nanoseconds={step.execution_time}
      />
    )
  if (step.status?.status === 'in-progress')
    return <RunningClock startedAt={step.started_at} label="" />
  return null
}

const StepRail = ({
  steps,
  selectedId,
  onSelect,
}: {
  steps: TWorkflowStep[]
  selectedId?: string
  onSelect: (id: string) => void
}) => (
  <nav aria-label="Workflow steps" className="flex flex-col gap-0.5">
    {steps.map((step) => (
      <Tooltip
        key={step.id}
        className="!w-full"
        position="right"
        tipContent={<StepTooltipContent step={step} steps={steps} />}
        tipContentClassName="!whitespace-normal"
      >
        <button
          type="button"
          aria-current={step.id === selectedId ? 'step' : undefined}
          className={cn(
            'flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-left cursor-pointer',
            'hover:bg-black/5 dark:hover:bg-white/10',
            'focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary-600',
            { 'bg-black/5 dark:bg-white/10': step.id === selectedId }
          )}
          onClick={() => onSelect(step.id ?? '')}
        >
          <span className="flex min-w-0 items-center gap-3">
            <Status
              isWithoutText
              status={step.status?.status || 'unknown'}
              variant="timeline"
            />
            <Text nowrap className="truncate">
              {getWorkflowStepTitle(step)}
            </Text>
          </span>
          <StepTiming step={step} />
        </button>
      </Tooltip>
    ))}
  </nav>
)

const ApprovalPrompt = ({
  step,
  isSelected,
  onApprove,
  onDeny,
  onView,
  onViewPlan,
}: {
  step: TWorkflowStep
  isSelected: boolean
  onApprove?: () => void
  onDeny?: () => void
  onView: () => void
  onViewPlan?: () => void
}) => {
  const copy =
    APPROVAL_BANNER_COPY[
      (step.approval?.type ??
        'terraform_plan') as keyof typeof APPROVAL_BANNER_COPY
    ]

  return (
    <Banner className="@container" theme="warn">
      <div className="flex flex-col gap-2">
        <div className="flex flex-col">
          <Text weight="strong">
            {isSelected
              ? copy.title
              : `${getWorkflowStepTitle(step)} is waiting for approval`}
          </Text>
          {isSelected ? (
            <Text variant="subtext" theme="neutral">
              {copy.message}
            </Text>
          ) : null}
        </div>

        <div className="flex self-end gap-2">
          {isSelected ? (
            <>
              {onViewPlan ? (
                <Button onClick={onViewPlan}>View plan</Button>
              ) : null}
              <Button onClick={onDeny}>Deny plan</Button>
              <Button variant="primary" onClick={onApprove}>
                Approve plan
              </Button>
            </>
          ) : (
            <Button onClick={onView}>View step</Button>
          )}
        </div>
      </div>
    </Banner>
  )
}

const StepLogs = ({
  logs,
  step,
}: {
  logs: TOTELLog[]
  step: TWorkflowStep
}) => (
  <LogStreamContext.Provider
    key={step.id}
    value={{
      logs,
      logStreamId: step.id ?? '',
      isLoading: false,
      isCatchingUp: false,
      error: null,
      connectionState: LIVE_STATUSES.has(step.status?.status ?? '')
        ? 'connected'
        : 'disconnected',
    }}
  >
    <LogViewerProvider>
      <SSELogs />
    </LogViewerProvider>
  </LogStreamContext.Provider>
)

export const RunView = ({
  steps,
  logs,
  onApprove,
  onDeny,
  renderPlan,
}: IRunView) => {
  const { addPanel, removePanel } = useSurfaces()
  const [picked, setPicked] = useState<string>()

  const current =
    steps.find((step) => CURRENT_STATUSES.has(step.status?.status ?? '')) ??
    steps[steps.length - 1]
  const selected = steps.find((step) => step.id === picked) ?? current
  const awaiting = steps.find(
    (step) => step.status?.status === 'approval-awaiting'
  )
  const stepLogs = (selected?.id && logs[selected.id]) || []
  const status = selected?.status?.status ?? 'pending'
  const empty = EMPTY_COPY[status] ?? EMPTY_COPY.pending

  return (
    <div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <StepRail
        steps={steps}
        selectedId={selected?.id}
        onSelect={(id) => setPicked(id === current?.id ? undefined : id)}
      />

      <div className="flex min-w-0 flex-col gap-4">
        {awaiting ? (
          <ApprovalPrompt
            step={awaiting}
            isSelected={awaiting.id === selected?.id}
            onApprove={onApprove}
            onDeny={onDeny}
            onView={() => setPicked(awaiting.id)}
            onViewPlan={
              renderPlan
                ? () =>
                    addPanel(
                      <Panel
                        size="3/4"
                        heading={`Plan for ${getWorkflowStepTitle(awaiting)}`}
                        footer={
                          <>
                            <Button
                              onClick={() => {
                                removePanel(PLAN_PANEL_ID)
                                onDeny?.()
                              }}
                            >
                              Deny plan
                            </Button>
                            <Button
                              variant="primary"
                              onClick={() => {
                                removePanel(PLAN_PANEL_ID)
                                onApprove?.()
                              }}
                            >
                              Approve plan
                            </Button>
                          </>
                        }
                      >
                        {renderPlan(awaiting)}
                      </Panel>,
                      undefined,
                      PLAN_PANEL_ID
                    )
                : undefined
            }
          />
        ) : null}

        {selected && !NOT_STARTED_STATUSES.has(status) ? (
          <StepLogs logs={stepLogs} step={selected} />
        ) : (
          <div className="flex h-[28rem] items-center rounded-md border">
            <EmptyState
              variant="table"
              emptyTitle={empty.title}
              emptyMessage={empty.message}
            />
          </div>
        )}
      </div>
    </div>
  )
}
