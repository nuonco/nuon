import { useEffect, useMemo, useState, useRef } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { completeDeployStep, getApp, getCurrentOnboarding, getInstall, getInstallStack, getRunnerLatestHeartbeat, getWorkflow, getWorkflowSteps } from '@/lib'
import { isRecentTimestamp } from '@/utils/time-utils'
import { getStatusTheme } from '@/utils/status-utils'
import { cn } from '@/utils/classnames'
import { toSentenceCase } from '@/utils/string-utils'
import type { TOnboarding, TWorkflow, TWorkflowStep } from '@/types'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

function useOnboardingWorkflow(onboarding: TOnboarding | undefined, setSharedData: (key: string, val: unknown) => void) {
  const orgId = onboarding?.org_id
  const workflowId = onboarding?.workflow_id

  const { data: polledOnboarding } = useQuery({
    queryKey: ['onboarding-provision-poll'],
    queryFn: getCurrentOnboarding,
    enabled: !!orgId && !workflowId,
    refetchInterval: 2000,
  })

  useEffect(() => {
    if (polledOnboarding?.workflow_id && !workflowId) {
      setSharedData('onboarding', polledOnboarding)
    }
  }, [polledOnboarding, workflowId, setSharedData])

  const effectiveWorkflowId = workflowId ?? polledOnboarding?.workflow_id

  const { data: workflow } = useQuery({
    queryKey: ['onboarding-workflow', effectiveWorkflowId],
    queryFn: () => getWorkflow({ workflowId: effectiveWorkflowId!, orgId: orgId! }),
    enabled: !!effectiveWorkflowId && !!orgId,
    refetchInterval: (query) => {
      const wf = query.state.data as TWorkflow | undefined
      return wf?.finished ? false : 4000
    },
  })

  const { data: steps = [] } = useQuery({
    queryKey: ['onboarding-workflow-steps', effectiveWorkflowId],
    queryFn: () => getWorkflowSteps({ workflowId: effectiveWorkflowId!, orgId: orgId! }),
    enabled: !!effectiveWorkflowId && !!orgId,
    refetchInterval: () => {
      return workflow?.finished ? false : 4000
    },
  })

  return { workflow, steps, workflowId: effectiveWorkflowId }
}

// --- Row model ---

interface IRow {
  id: string
  label: string
  description: string
  errorDescription?: string
  errorStepId?: string
  status: 'pending' | 'active' | 'done' | 'error'
  completed: number
  total: number
}

const HIDDEN_STEPS = new Set(['provision runner service account', 'generate install state', 'await runner healthy'])
const RUNNER_STEPS = new Set(['await runner health'])
const STACK_STEPS = new Set(['generate install stack', 'await install stack', 'update install stack outputs'])
const SANDBOX_STEPS = new Set(['provision sandbox plan', 'provision sandbox apply plan', 'sync secrets', 'provision sandbox dns if enabled'])

function getComponentName(stepName: string): string | null {
  const syncMatch = stepName.match(/^sync and plan (.+)$/)
  if (syncMatch) return syncMatch[1]
  const applyMatch = stepName.match(/^apply (.+)$/)
  if (applyMatch) return applyMatch[1]
  const syncImageMatch = stepName.match(/^sync (.+)$/)
  if (syncImageMatch) return syncImageMatch[1]
  return null
}

function resolveStatus(steps: TWorkflowStep[]): IRow['status'] {
  if (steps.length === 0) return 'pending'
  if (steps.some((s) => getStatusTheme(s.status?.status || 'pending') === 'error')) return 'error'
  if (steps.every((s) => getStatusTheme(s.status?.status || 'pending') === 'success')) return 'done'
  if (steps.some((s) => {
    const st = s.status?.status || 'pending'
    return st !== 'pending' && st !== 'success'
  })) return 'active'
  return 'pending'
}

function countDone(steps: TWorkflowStep[]): number {
  return steps.filter((s) => getStatusTheme(s.status?.status || 'pending') === 'success').length
}

function isHumanReadable(desc: string): boolean {
  if (desc.length < 6) return false
  if (/^[a-z]+ \d+$/i.test(desc)) return false
  if (!/[a-zA-Z]{3,}/.test(desc)) return false
  return true
}

function getStepDescription(steps: TWorkflowStep[]): string | undefined {
  const inFlight = steps.find((s) => {
    const st = s.status?.status || 'pending'
    return st !== 'pending' && getStatusTheme(st) !== 'success'
  })
  const desc = inFlight?.status?.status_human_description
  if (desc && isHumanReadable(desc)) return desc

  for (let i = steps.length - 1; i >= 0; i--) {
    const d = steps[i].status?.status_human_description
    if (d && isHumanReadable(d)) return d
  }
  return undefined
}

function getErrorInfo(steps: TWorkflowStep[]): { description?: string; stepId?: string } {
  const errorStep = steps.find((s) => getStatusTheme(s.status?.status || 'pending') === 'error')
  if (!errorStep) return {}
  const desc = errorStep.status?.status_human_description
  return {
    description: desc && isHumanReadable(desc) ? desc : undefined,
    stepId: errorStep.id,
  }
}

const FALLBACK_COPY: Record<string, Record<IRow['status'], string>> = {
  stack: { pending: 'Waiting to provision...', active: 'Provisioning stack...', done: 'Stack provisioned', error: 'Stack failed' },
  runner: { pending: 'Waiting to start...', active: 'Awaiting health check...', done: 'Healthy', error: 'Runner failed' },
  sandbox: { pending: 'Waiting to configure...', active: 'Setting up your sandbox...', done: 'Sandbox ready', error: 'Sandbox failed' },
}

function useProvisioningRows(steps: TWorkflowStep[]): IRow[] {
  return useMemo(() => {
    const runnerSteps: TWorkflowStep[] = []
    const stackSteps: TWorkflowStep[] = []
    const sandboxSteps: TWorkflowStep[] = []
    const componentMap = new Map<string, TWorkflowStep[]>()

    for (const step of steps) {
      if (step.execution_type === 'hidden' || HIDDEN_STEPS.has(step.name)) continue

      if (RUNNER_STEPS.has(step.name)) {
        runnerSteps.push(step)
      } else if (STACK_STEPS.has(step.name)) {
        stackSteps.push(step)
      } else if (SANDBOX_STEPS.has(step.name)) {
        sandboxSteps.push(step)
      } else {
        const compName = getComponentName(step.name)
        if (compName) {
          if (!componentMap.has(compName)) componentMap.set(compName, [])
          componentMap.get(compName)!.push(step)
        }
      }
    }

    const rows: IRow[] = []

    const addInfraRow = (id: string, label: string, bucket: TWorkflowStep[]) => {
      const status = resolveStatus(bucket)
      const fallback = FALLBACK_COPY[id]?.[status] ?? ''
      const apiDesc = status === 'error' ? undefined : getStepDescription(bucket)
      const errorInfo = status === 'error' ? getErrorInfo(bucket) : {}
      rows.push({
        id,
        label,
        description: apiDesc || fallback,
        errorDescription: errorInfo.description,
        errorStepId: errorInfo.stepId,
        status,
        completed: countDone(bucket),
        total: bucket.length,
      })
    }

    addInfraRow('stack', 'CloudFormation Stack', stackSteps)
    addInfraRow('runner', 'Runner', runnerSteps)
    addInfraRow('sandbox', 'Sandbox', sandboxSteps)

    for (const [compName, compSteps] of componentMap) {
      const status = resolveStatus(compSteps)
      const fallback: Record<IRow['status'], string> = {
        pending: 'Waiting to deploy...',
        active: `Deploying ${compName}...`,
        done: 'Deployed',
        error: 'Failed',
      }
      const apiDesc = status === 'error' ? undefined : getStepDescription(compSteps)
      const errorInfo = status === 'error' ? getErrorInfo(compSteps) : {}
      rows.push({
        id: `component-${compName}`,
        label: toTitleCase(compName),
        description: apiDesc || fallback[status],
        errorDescription: errorInfo.description,
        errorStepId: errorInfo.stepId,
        status,
        completed: countDone(compSteps),
        total: compSteps.length,
      })
    }

    return rows
  }, [steps])
}

function toTitleCase(str: string): string {
  return str
    .replace(/[-_]/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

// --- ETA ---

const AVG_SECONDS_PER_STEP = 90

function useEta(doneCount: number, totalCount: number, isActive: boolean) {
  const [elapsed, setElapsed] = useState(0)
  const startRef = useRef(Date.now())

  useEffect(() => {
    if (!isActive) {
      setElapsed(0)
      startRef.current = Date.now()
      return
    }
    startRef.current = Date.now()
    const id = setInterval(() => {
      setElapsed(Math.floor((Date.now() - startRef.current) / 1000))
    }, 1000)
    return () => clearInterval(id)
  }, [isActive])

  if (!isActive || totalCount === 0) return null

  const remaining = totalCount - doneCount
  if (remaining <= 0) return null

  if (doneCount > 0 && elapsed > 10) {
    const secondsPerStep = elapsed / doneCount
    const etaSeconds = Math.max(0, Math.round(remaining * secondsPerStep))
    return formatEta(etaSeconds)
  }

  return formatEta(remaining * AVG_SECONDS_PER_STEP)
}

function formatEta(seconds: number): string {
  if (seconds < 60) return '< 1 min'
  const m = Math.ceil(seconds / 60)
  return `~${m} min`
}

// --- Runner metadata ---

interface IRunnerMeta {
  version?: string
  connected: boolean
}

function useRunnerMeta(orgId?: string, installId?: string, runnerDone?: boolean): IRunnerMeta | undefined {
  const { data: install } = useQuery({
    queryKey: ['onboarding-install', installId],
    queryFn: () => getInstall({ installId: installId!, orgId: orgId! }),
    enabled: !!installId && !!orgId && !!runnerDone,
  })

  const runnerId = install?.runner_id

  const { data: heartbeats } = useQuery({
    queryKey: ['onboarding-runner-heartbeat', runnerId],
    queryFn: () => getRunnerLatestHeartbeat({ runnerId: runnerId!, orgId: orgId! }),
    enabled: !!runnerId && !!orgId,
  })

  if (!runnerId || !runnerDone) return undefined

  const heartbeat = heartbeats?.install ?? heartbeats?.org ?? heartbeats?.[''] ?? undefined

  return {
    version: heartbeat?.version,
    connected: isRecentTimestamp(heartbeat?.created_at),
  }
}

// --- Progress Ring ---

const ROW_RING_STROKE = 2.5
const APP_RING_STROKE = 5
const ROW_RING_SIZE = 24
const APP_RING_SIZE = 44

interface IProgressRingProps {
  completed: number
  total: number
  status: IRow['status']
  size?: number
  strokeWidth?: number
  showPercentage?: boolean
}

function ProgressRing({
  completed,
  total,
  status,
  size = ROW_RING_SIZE,
  strokeWidth = ROW_RING_STROKE,
  showPercentage = false,
}: IProgressRingProps) {
  const radius = (size - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const ratio = total > 0 ? completed / total : 0
  const offset = circumference * (1 - ratio)
  const pct = Math.round(ratio * 100)
  const isAppSize = size >= APP_RING_SIZE

  if (status === 'done') {
    return (
      <div className="flex items-center justify-center shrink-0 animate-[scale-in_0.3s_ease-out]" style={{ width: size, height: size }}>
        <Icon variant="CheckCircleIcon" size={size} weight="fill" theme="success" />
      </div>
    )
  }

  if (status === 'error') {
    return (
      <div className="flex items-center justify-center shrink-0" style={{ width: size, height: size }}>
        <Icon variant="XCircleIcon" size={size} weight="fill" theme="error" />
      </div>
    )
  }

  if (status === 'pending') {
    return (
      <div className="flex items-center justify-center shrink-0" style={{ width: size, height: size }}>
        <svg width={size} height={size}>
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            strokeWidth={strokeWidth}
            style={{ stroke: 'var(--border-color)' }}
          />
        </svg>
      </div>
    )
  }

  if (isAppSize && ratio > 0) {
    return (
      <div className="relative flex items-center justify-center shrink-0" style={{ width: size, height: size }}>
        <svg width={size} height={size} className="-rotate-90">
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            strokeWidth={strokeWidth}
            style={{ stroke: 'var(--border-color)' }}
          />
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            strokeWidth={strokeWidth}
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            className="transition-[stroke-dashoffset] duration-700 ease-out"
            style={{ stroke: 'var(--color-green-600)' }}
          />
        </svg>
        {showPercentage && pct > 0 && (
          <span className="absolute text-xs font-semibold leading-none" style={{ color: 'var(--foreground)' }}>
            {pct}%
          </span>
        )}
      </div>
    )
  }

  const arcLen = circumference * 0.25

  return (
    <div className="relative flex items-center justify-center shrink-0" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          strokeWidth={strokeWidth}
          style={{ stroke: 'var(--border-color)' }}
        />
      </svg>
      <svg
        width={size}
        height={size}
        className="absolute inset-0"
        style={{ transformOrigin: 'center', animation: 'spinner-rotate 1s linear infinite' }}
      >
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          strokeDasharray={`${arcLen} ${circumference - arcLen}`}
          style={{ stroke: 'var(--color-green-600)' }}
        />
      </svg>
    </div>
  )
}

// --- Install stack quick link ---

function useInstallStackQuickLink(orgId?: string, installId?: string) {
  const { data: stack } = useQuery({
    queryKey: ['onboarding-install-stack', installId],
    queryFn: () => getInstallStack({ installId: installId!, orgId: orgId! }),
    enabled: !!installId && !!orgId,
    refetchInterval: (query) =>
      query.state.data?.versions?.[0]?.quick_link_url ? false : 3000,
  })

  return stack?.versions?.[0]?.quick_link_url ?? undefined
}

// --- Row component ---

function ProvisioningRow({ row, isLast, runnerMeta, quickLinkUrl, workflowUrl }: { row: IRow; isLast: boolean; runnerMeta?: IRunnerMeta; quickLinkUrl?: string; workflowUrl?: string }) {
  const isActive = row.status === 'active'
  const isRunnerDone = row.id === 'runner' && row.status === 'done' && runnerMeta

  return (
    <div
      className={cn(
        'flex items-center gap-4 px-5 py-4 transition-colors duration-300',
        !isLast && 'border-b',
        isActive && 'bg-cool-grey-50/50 dark:bg-dark-grey-50/30',
      )}
    >
      <div className="flex flex-col flex-1 min-w-0 gap-0">
        <Text variant="base" weight="strong">{row.label}</Text>
        {isRunnerDone ? (
          <div className="flex items-center gap-4 mt-0.5">
            <Text variant="body" className="text-green-700 dark:text-green-500">
              Healthy
            </Text>
            <div className="flex items-center gap-1">
              <Text variant="body" className="text-cool-grey-500 dark:text-cool-grey-500">Status:</Text>
              <Text variant="body" className={runnerMeta.connected ? 'text-green-700 dark:text-green-500' : 'text-cool-grey-600 dark:text-cool-grey-400'}>
                {runnerMeta.connected ? 'Connected' : 'Disconnected'}
              </Text>
            </div>
            {runnerMeta.version && (
              <div className="flex items-center gap-1">
                <Text variant="body" className="text-cool-grey-500 dark:text-cool-grey-500">Version:</Text>
                <Text variant="body" className="text-cool-grey-700 dark:text-cool-grey-300">
                  {runnerMeta.version}
                </Text>
              </div>
            )}
          </div>
        ) : (
          <Text
            variant="body"
            theme={row.status === 'error' && row.errorDescription
              ? 'error'
              : row.status === 'done' || row.status === 'active'
                ? 'success'
                : 'neutral'}
          >
            {row.status === 'error' && row.errorDescription
              ? row.errorDescription
              : row.description}
          </Text>
        )}
      </div>
      {quickLinkUrl && row.status !== 'done' && (
        <Button
          type="button"
          variant="secondary"
          size="sm"
          className="shrink-0"
          onClick={() => window.open(quickLinkUrl, '_blank', 'noopener,noreferrer')}
        >
          Launch in AWS <Icon variant="ArrowSquareOutIcon" size={14} />
        </Button>
      )}
      {row.status === 'error' && workflowUrl && (
        <Button
          href={workflowUrl}
          variant="secondary"
          size="sm"
          className="shrink-0"
        >
          View details
        </Button>
      )}
      <ProgressRing completed={row.completed} total={row.total} status={row.status} size={ROW_RING_SIZE} />
    </div>
  )
}

// --- Main step ---

export const ProvisioningStepContainer = ({
  onAdvance,
  onGoBack,
  sharedData,
  setSharedData,
}: IWizardStepComponentProps) => {
  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id
  const appId = onboarding?.app_id
  const cloudProvider = (onboarding?.cloud_provider as 'aws' | 'gcp' | 'azure') || 'aws'

  const { data: app } = useQuery({
    queryKey: ['app', appId],
    queryFn: () => getApp({ appId: appId!, orgId: orgId! }),
    enabled: !!appId && !!orgId,
  })

  const { workflow, steps, workflowId } = useOnboardingWorkflow(onboarding, setSharedData)

  const isFinished = !!workflow?.finished
  const isError = workflow?.status?.status === 'error'
  const isProvisioning = !!workflowId && !isFinished && !isError

  const rows = useProvisioningRows(steps)
  const doneCount = rows.filter((r) => r.status === 'done').length
  const eta = useEta(doneCount, rows.length, isProvisioning)

  const runnerRow = rows.find((r) => r.id === 'runner')
  const runnerMeta = useRunnerMeta(orgId, onboarding?.install_id, runnerRow?.status === 'done')
  const isCloudInstall = onboarding?.install_mode === 'cloud'
  const quickLinkUrl = useInstallStackQuickLink(
    isCloudInstall ? orgId : undefined,
    isCloudInstall ? onboarding?.install_id : undefined,
  )

  const activeRow = rows.find((r) => r.status === 'active')
  const nextPendingRow = rows.find((r) => r.status === 'pending')
  const allRowsDone = rows.length > 0 && doneCount === rows.length
  const dynamicMessage = activeRow
    ? `${activeRow.label}: ${activeRow.description}`
    : nextPendingRow
      ? `Up next: ${nextPendingRow.label}`
      : allRowsDone && !isFinished
        ? 'Finishing up...'
        : 'Preparing your environment...'

  const { mutate: completeDeploy, isPending: deployPending } = useMutation({
    mutationFn: () => completeDeployStep({ orgId: orgId! }),
    onSuccess: (ob) => {
      setSharedData('onboarding', ob)
      onAdvance()
    },
  })

  const appName = app?.name ?? onboarding?.example_app_slug ?? 'Your app'
  const appDone = isFinished && !isError

  return (
    <div className="flex flex-col gap-8">
      {isError && workflow?.status?.status_human_description && (
        <Banner theme="error">{toSentenceCase(workflow.status.status_human_description)}</Banner>
      )}

      {/* App card */}
      <Card className="!gap-0 !p-4">
        <div className="flex items-center gap-4">
          <CloudPlatform platform={cloudProvider} colorVariant="color" displayVariant="icon-only" iconSize="24" className="shrink-0" />
          <div className="flex flex-col flex-1 min-w-0 gap-0">
            <Text variant="base" weight="strong" className="truncate">{appName}</Text>
            <Text variant="body" className="text-cool-grey-600 dark:text-cool-grey-400">
              {appDone ? 'Deployed' : dynamicMessage}
            </Text>
          </div>
          {appDone ? (
            <div className="flex items-center gap-2 shrink-0">
              <Icon variant="CheckCircleIcon" size={20} weight="fill" theme="success" />
              <Text variant="body" weight="strong" theme="success">All resources provisioned</Text>
            </div>
          ) : (
            <ProgressRing
              completed={doneCount}
              total={allRowsDone && !isFinished ? rows.length + 1 : rows.length || 1}
              status="active"
              size={APP_RING_SIZE}
              strokeWidth={APP_RING_STROKE}
              showPercentage
            />
          )}
        </div>
      </Card>

      {/* Install rows */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <Text variant="base" weight="strong">
            {isError ? 'Provisioning failed' : 'Provisioning resources'}
          </Text>
          {!appDone && !isError && eta && (
            <Text variant="body" theme="neutral">
              ETA {eta}
            </Text>
          )}
        </div>
        <Card className="!p-0 !gap-0 overflow-hidden">
          {rows.length === 0 ? (
            ['CloudFormation Stack', 'Runner', 'Sandbox'].map((label, i) => (
              <div key={label} className={cn('flex items-center gap-4 px-5 py-4', i < 2 && 'border-b')}>
                <div className="flex flex-col flex-1 min-w-0 gap-0.5">
                  <Text weight="strong" theme="neutral">{label}</Text>
                  <Text variant="subtext" theme="neutral">Waiting to start...</Text>
                </div>
                <ProgressRing completed={0} total={0} status="pending" size={ROW_RING_SIZE} />
              </div>
            ))
          ) : (
            <>
              {rows.map((row, i) => {
                const hasMore = allRowsDone && !isFinished
                return (
                  <ProvisioningRow
                    key={row.id}
                    row={row}
                    isLast={!hasMore && i === rows.length - 1}
                    runnerMeta={row.id === 'runner' ? runnerMeta : undefined}
                    quickLinkUrl={row.id === 'stack' ? quickLinkUrl : undefined}
                    workflowUrl={row.status === 'error' && row.errorStepId && orgId && onboarding?.install_id && workflowId
                      ? `/${orgId}/installs/${onboarding.install_id}/workflows/${workflowId}?panel=${row.errorStepId}`
                      : undefined}
                  />
                )
              })}
              {allRowsDone && !isFinished && (
                <div className="flex items-center gap-4 px-5 py-4">
                  <div className="flex flex-col flex-1 min-w-0 gap-0">
                    <Text variant="base" weight="strong">Deploying components</Text>
                    <Text variant="body" className="text-blue-700 dark:text-blue-500">Preparing deployments...</Text>
                  </div>
                  <ProgressRing completed={0} total={0} status="active" size={ROW_RING_SIZE} />
                </div>
              )}
            </>
          )}
        </Card>
      </div>

      {/* Navigation */}
      <div className="flex justify-between">
        {onGoBack ? (
          <Button type="button" variant="secondary" onClick={onGoBack}>
            <Icon variant="CaretLeftIcon" weight="bold" /> Back
          </Button>
        ) : (
          <div />
        )}
        <Button
          type="button"
          variant="primary"
          disabled={!isFinished || isError || deployPending}
          onClick={() => completeDeploy()}
        >
          {deployPending ? 'Completing...' : 'Continue'} {!deployPending && <Icon variant="CaretRightIcon" weight="bold" />}
        </Button>
      </div>
    </div>
  )
}
