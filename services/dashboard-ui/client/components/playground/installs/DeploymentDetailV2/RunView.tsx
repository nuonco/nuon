import { useEffect, useRef, useState, type ReactNode } from 'react'
import { APPROVAL_BANNER_COPY } from '@/components/approvals/ApprovalBanner'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
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
import type { IRunAttempt } from './run-simulation'
import { planStepFor, StepTooltipContent, stepStatusLabel } from './StepTooltip'
import {
  getPolicyViolationCounts,
  getStepBanner,
  getWorkflowStepTitle,
} from '@/utils/workflow-utils'

const CURRENT_STATUSES = new Set([
  'in-progress',
  'approval-awaiting',
  'approval-denied',
  'failed-pending-retry',
  'error',
])

const ATTEMPT_KEY_PREFIX = 'attempt'

const attemptKey = (number: number) => `${ATTEMPT_KEY_PREFIX}${number}`

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
  attempts?: Record<string, IRunAttempt[]>
  autoApprove?: boolean
  onApprove?: () => void
  onDeny?: () => void
  onRetry?: () => void
  onSkip?: () => void
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

const attemptsOf = (
  step: TWorkflowStep,
  prior: IRunAttempt[],
  logs: TOTELLog[]
): IRunAttempt[] => [
  ...prior,
  {
    number: prior.length + 1,
    status: step.status?.status ?? 'pending',
    description: step.status?.status_human_description,
    startedAt: step.started_at,
    executionNs: step.execution_time,
    logs,
  },
]

const StepAttemptLabel = ({
  step,
  attemptCount,
}: {
  step: TWorkflowStep
  attemptCount: number
}) => {
  if (step.status?.status === 'failed-pending-retry')
    return (
      <Text variant="subtext" weight="strong" theme="warn" nowrap>
        {step.retryable ? 'Needs action' : 'Retrying'}
      </Text>
    )
  if (step.status?.status === 'user-skipped')
    return (
      <Text variant="subtext" theme="neutral" nowrap>
        Skipped
      </Text>
    )
  if (attemptCount > 1)
    return (
      <Text variant="subtext" theme="neutral" nowrap>
        Attempt {attemptCount}
      </Text>
    )
  return null
}

const StepSubline = ({ step }: { step: TWorkflowStep }) => {
  const { denyCount, warnCount, skipFindings } = getPolicyViolationCounts(step)
  const segments = [
    <Status key="status" status={step.status?.status ?? 'unknown'}>
      {stepStatusLabel(step)}
    </Status>,
    step.finished || step.status?.status === 'in-progress' ? (
      <StepTiming key="timing" step={step} />
    ) : null,
    denyCount > 0 ? (
      <Text key="deny" variant="subtext" weight="strong" theme="error" nowrap>
        {denyCount} {denyCount === 1 ? 'violation' : 'violations'}
      </Text>
    ) : null,
    warnCount > 0 ? (
      <Text key="warn" variant="subtext" weight="strong" theme="warn" nowrap>
        {warnCount} {warnCount === 1 ? 'warning' : 'warnings'}
      </Text>
    ) : null,
    skipFindings.length > 0 ? (
      <Text key="skip" variant="subtext" theme="neutral" nowrap>
        {skipFindings.length} skipped
      </Text>
    ) : null,
  ].filter(Boolean)

  return (
    <span className="flex flex-wrap items-center gap-x-1.5">
      {segments.flatMap((segment, index) =>
        index === 0
          ? [segment]
          : [
              <Text key={`dot-${index}`} variant="subtext" theme="neutral">
                ·
              </Text>,
              segment,
            ]
      )}
    </span>
  )
}

const StepRail = ({
  steps,
  attemptsByStep,
  selectedId,
  onSelect,
}: {
  steps: TWorkflowStep[]
  attemptsByStep: Record<string, IRunAttempt[]>
  selectedId?: string
  onSelect: (id: string) => void
}) => {
  const navRef = useRef<HTMLElement>(null)

  useEffect(() => {
    const nav = navRef.current
    const current = nav?.querySelector<HTMLElement>('[aria-current="step"]')
    if (!nav || !current) return

    const offset =
      current.getBoundingClientRect().top - nav.getBoundingClientRect().top
    if (offset < 0) nav.scrollTop += offset
    else if (offset + current.offsetHeight > nav.clientHeight)
      nav.scrollTop += offset + current.offsetHeight - nav.clientHeight
  }, [selectedId])

  return (
    <nav
      ref={navRef}
      aria-label="Workflow steps"
      className="flex flex-col gap-0.5 lg:sticky lg:top-0 lg:max-h-screen lg:overflow-y-auto"
    >
      {steps.map((step) => (
        <Tooltip
          key={step.id}
          className="!w-full"
          position="right"
          tipContent={
            <StepTooltipContent
              step={step}
              steps={steps}
              attempts={attemptsByStep[step.id ?? '']}
            />
          }
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
            <span className="flex shrink-0 items-center gap-2">
              <StepAttemptLabel
                step={step}
                attemptCount={attemptsByStep[step.id ?? '']?.length ?? 1}
              />
              <StepTiming step={step} />
            </span>
          </button>
        </Tooltip>
      ))}
    </nav>
  )
}

const FailedStepPrompt = ({
  step,
  isSelected,
  onRetry,
  onSkip,
  onView,
}: {
  step: TWorkflowStep
  isSelected: boolean
  onRetry?: () => void
  onSkip?: () => void
  onView: () => void
}) => {
  const banner = getStepBanner(step)

  return (
    <Banner className="@container" theme="error">
      <div className="flex flex-col gap-2">
        <div className="flex flex-col">
          <Text weight="strong">
            {isSelected
              ? banner?.title
              : `${getWorkflowStepTitle(step)} needs your attention`}
          </Text>
          {isSelected ? (
            <Text variant="subtext" theme="neutral">
              {banner?.copy}
            </Text>
          ) : null}
        </div>

        <div className="flex self-end gap-2">
          {isSelected ? (
            <>
              <Button variant="danger" onClick={onSkip}>
                Skip step
              </Button>
              <Button variant="primary" onClick={onRetry}>
                Retry step
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
  attempt,
  isLatest,
  logs,
  step,
}: {
  attempt: number
  isLatest: boolean
  logs: TOTELLog[]
  step: TWorkflowStep
}) => (
  <LogStreamContext.Provider
    key={`${step.id}-${attempt}`}
    value={{
      logs,
      logStreamId: `${step.id}-${attempt}`,
      isLoading: false,
      isCatchingUp: false,
      error: null,
      connectionState:
        isLatest && LIVE_STATUSES.has(step.status?.status ?? '')
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
  attempts = {},
  autoApprove = false,
  onApprove,
  onDeny,
  onRetry,
  onSkip,
  renderPlan,
}: IRunView) => {
  const { addPanel, removePanel } = useSurfaces()
  const [picked, setPicked] = useState<string>()
  const [pickedAttempts, setPickedAttempts] = useState<
    Record<string, number | undefined>
  >({})

  const current =
    steps.find((step) => CURRENT_STATUSES.has(step.status?.status ?? '')) ??
    steps[steps.length - 1]
  const selected = steps.find((step) => step.id === picked) ?? current
  const awaiting = steps.find(
    (step) => step.status?.status === 'approval-awaiting'
  )
  const openPlan = (planStep: TWorkflowStep) => {
    const canRespond =
      planStep.status?.status === 'approval-awaiting' && !autoApprove

    addPanel(
      <Panel
        size="3/4"
        heading={`Plan for ${getWorkflowStepTitle(planStep)}`}
        footer={
          canRespond ? (
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
          ) : undefined
        }
      >
        {renderPlan?.(planStep)}
      </Panel>,
      undefined,
      PLAN_PANEL_ID
    )
  }
  const selectedPlan = selected ? planStepFor(selected, steps) : undefined
  const failed = steps.find(
    (step) => step.status?.status === 'failed-pending-retry' && step.retryable
  )
  const skippedBanner =
    selected?.status?.status === 'user-skipped'
      ? getStepBanner(selected)
      : undefined
  const attemptsByStep = Object.fromEntries(
    steps.map((step) => [
      step.id ?? '',
      attemptsOf(
        step,
        attempts[step.id ?? ''] ?? [],
        logs[step.id ?? ''] ?? []
      ),
    ])
  )
  const selectedAttempts = attemptsByStep[selected?.id ?? ''] ?? []
  const latestAttempt = selectedAttempts.at(-1)
  const attempt =
    selectedAttempts.find(
      (item) => item.number === pickedAttempts[selected?.id ?? '']
    ) ?? latestAttempt
  const status = selected?.status?.status ?? 'pending'
  const empty = EMPTY_COPY[status] ?? EMPTY_COPY.pending

  return (
    <div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <StepRail
        steps={steps}
        attemptsByStep={attemptsByStep}
        selectedId={selected?.id}
        onSelect={(id) => setPicked(id === current?.id ? undefined : id)}
      />

      <div className="flex min-w-0 flex-col gap-4">
        {awaiting && !autoApprove ? (
          <ApprovalPrompt
            step={awaiting}
            isSelected={awaiting.id === selected?.id}
            onApprove={onApprove}
            onDeny={onDeny}
            onView={() => setPicked(awaiting.id)}
            onViewPlan={renderPlan ? () => openPlan(awaiting) : undefined}
          />
        ) : null}

        {selected ? (
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-col gap-1">
              <Text variant="h3" weight="strong">
                {getWorkflowStepTitle(selected)}
              </Text>
              <StepSubline step={selected} />
            </div>
            {selectedPlan && renderPlan ? (
              <Button onClick={() => openPlan(selectedPlan)}>View plan</Button>
            ) : null}
          </div>
        ) : null}

        {failed ? (
          <FailedStepPrompt
            step={failed}
            isSelected={failed.id === selected?.id}
            onRetry={onRetry}
            onSkip={onSkip}
            onView={() => setPicked(failed.id)}
          />
        ) : null}

        {skippedBanner ? (
          <Banner theme={skippedBanner.theme}>
            <div className="flex flex-col">
              <Text weight="strong">{skippedBanner.title}</Text>
              <Text variant="subtext" theme="neutral">
                {skippedBanner.copy}
              </Text>
            </div>
          </Banner>
        ) : null}

        {selected && attempt && !NOT_STARTED_STATUSES.has(status) ? (
          selectedAttempts.length > 1 ? (
            <Tabs
              naturalHeight
              activeTab={attemptKey(attempt.number)}
              onTabChange={(key) =>
                setPickedAttempts((current) => ({
                  ...current,
                  [selected.id ?? '']:
                    key === attemptKey(latestAttempt?.number ?? 0)
                      ? undefined
                      : Number(key.replace(ATTEMPT_KEY_PREFIX, '')),
                }))
              }
              tabLabelNodes={Object.fromEntries(
                selectedAttempts.map((item) => [
                  attemptKey(item.number),
                  <span key={item.number} className="flex items-center gap-2">
                    <Status
                      isWithoutText
                      iconSize={14}
                      status={item.status}
                      variant="timeline"
                    />
                    Attempt {item.number}
                  </span>,
                ])
              )}
              tabs={Object.fromEntries(
                selectedAttempts.map((item) => [
                  attemptKey(item.number),
                  <div key={item.number} className="pt-4">
                    <StepLogs
                      attempt={item.number}
                      isLatest={item.number === latestAttempt?.number}
                      logs={item.logs}
                      step={selected}
                    />
                  </div>,
                ])
              )}
            />
          ) : (
            <StepLogs
              attempt={attempt.number}
              isLatest
              logs={attempt.logs}
              step={selected}
            />
          )
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
