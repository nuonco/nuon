import { Fragment, useEffect, useMemo } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Time } from '@/components/common/Time'
import { completeDeployStep, getCurrentOnboarding, getInstall, getRunner, getRunnerLatestHeartbeat, getWorkflow, getWorkflowSteps } from '@/lib'
import { getStepBadge } from '@/utils/workflow-utils'
import { toSentenceCase } from '@/utils/string-utils'
import { getStatusTheme, getStatusIconVariant } from '@/utils/status-utils'
import { cn } from '@/utils/classnames'
import type { TCloudPlatform, TOnboarding, TWorkflow, TWorkflowStep } from '@/types'
import { isLessThan30SecondsOld } from '@/utils/time-utils'
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
    queryFn: () => getWorkflowSteps({ workflowId: effectiveWorkflowId!, orgId: orgId!, limit: 100, offset: 0 }),
    enabled: !!effectiveWorkflowId && !!orgId,
    refetchInterval: (query) => {
      return workflow?.finished ? false : 4000
    },
  })

  return { workflow, steps, workflowId: effectiveWorkflowId }
}

const TIMELINE_NODE_CLASSES: Record<string, string> = {
  default: 'bg-cool-grey-200 dark:bg-cool-grey-800 dark:text-cool-grey-400',
  neutral: 'bg-cool-grey-200 dark:bg-cool-grey-800 dark:text-cool-grey-400',
  success: 'bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-400',
  error: 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-400',
  warn: 'bg-orange-100 text-orange-800 dark:bg-orange-950 dark:text-orange-400',
  info: 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-400',
  brand: 'bg-primary-200 text-primary-800 dark:bg-primary-950 dark:text-primary-400',
}

const HIDDEN_STEPS = new Set(['provision runner service account', 'generate install state', 'await runner healthy'])

const INSTALL_STACK_STEPS = new Set([
  'generate install stack',
  'await install stack',
  'update install stack outputs',
])

const SANDBOX_STEPS = new Set([
  'provision sandbox plan',
  'provision sandbox apply plan',
  'sync secrets',
  'provision sandbox dns if enabled',
])

interface IDisplayItem {
  id: string
  label: string
  status: string
  steps: TWorkflowStep[]
  group?: IStepGroup
}

function getGroupStatus(steps: TWorkflowStep[]): string {
  if (steps.some((s) => s.status?.status === 'error')) return 'error'
  if (steps.every((s) => s.status?.status === 'success')) return 'success'
  if (steps.some((s) => s.status?.status && s.status.status !== 'pending' && s.status.status !== 'success')) {
    return steps.find((s) => s.status?.status && s.status.status !== 'pending' && s.status.status !== 'success')!.status!.status!
  }
  return 'pending'
}

interface IStepGroup {
  match: Set<string>
  id: string
  label: string
  stepLabels?: Record<string, string>
}

const STEP_GROUPS: IStepGroup[] = [
  {
    match: INSTALL_STACK_STEPS,
    id: 'group-install-stack',
    label: 'Create stack',
    stepLabels: {
      'generate install stack': 'Generate',
      'await install stack': 'Await',
      'update install stack outputs': 'Update',
    },
  },
  {
    match: SANDBOX_STEPS,
    id: 'group-sandbox',
    label: 'Provision sandbox',
    stepLabels: {
      'provision sandbox plan': 'Plan',
      'provision sandbox apply plan': 'Apply',
      'sync secrets': 'Sync secrets',
      'provision sandbox dns if enabled': 'Provision dns',
    },
  },
]

function getComponentName(stepName: string): string | null {
  const syncMatch = stepName.match(/^sync and plan (.+)$/)
  if (syncMatch) return syncMatch[1]
  const applyMatch = stepName.match(/^apply (.+)$/)
  if (applyMatch) return applyMatch[1]
  const syncImageMatch = stepName.match(/^sync (.+)$/)
  if (syncImageMatch) return syncImageMatch[1]
  return null
}

function isComponentActionStep(stepName: string): boolean {
  return /\(pre_deploy_component\)$/.test(stepName) || /\(post_deploy_component\)$/.test(stepName)
}

function getActionStepLabel(stepName: string): string {
  const actionName = stepName.replace(/\s*Action Run\s*\(.*\)$/, '')
  return actionName
}

function getStepGroup(name: string): IStepGroup | null {
  const staticGroup = STEP_GROUPS.find((g) => g.match.has(name))
  if (staticGroup) return staticGroup

  const componentName = getComponentName(name)
  if (componentName) {
    return {
      match: new Set(),
      id: `group-component-${componentName}`,
      label: `Deploy ${componentName}`,
      stepLabels: {
        [`sync and plan ${componentName}`]: 'Plan',
        [`apply ${componentName}`]: 'Apply',
        [`sync ${componentName}`]: 'Sync',
      },
    }
  }

  return null
}

function findAdjacentComponentGroup(steps: TWorkflowStep[], currentIdx: number, direction: 'before' | 'after'): IStepGroup | null {
  const delta = direction === 'before' ? 1 : -1
  for (let i = currentIdx + delta; i >= 0 && i < steps.length; i += delta) {
    const step = steps[i]
    if (step.execution_type === 'hidden' || HIDDEN_STEPS.has(step.name)) continue
    const group = getComponentName(step.name) ? getStepGroup(step.name) : null
    if (group) return group
    if (!isComponentActionStep(step.name)) break
  }
  return null
}

function useDisplayItems(steps: TWorkflowStep[]): IDisplayItem[] {
  return useMemo(() => {
    const items: IDisplayItem[] = []
    let activeGroup: { config: IStepGroup; steps: TWorkflowStep[] } | null = null

    const flushGroup = () => {
      if (!activeGroup || activeGroup.steps.length === 0) return
      items.push({
        id: activeGroup.config.id,
        label: activeGroup.config.label,
        status: getGroupStatus(activeGroup.steps),
        steps: [...activeGroup.steps],
        group: activeGroup.config,
      })
      activeGroup = null
    }

    for (let i = 0; i < steps.length; i++) {
      const step = steps[i]
      if (step.execution_type === 'hidden') continue
      if (HIDDEN_STEPS.has(step.name)) continue

      const group = getStepGroup(step.name)

      if (group) {
        if (activeGroup && activeGroup.config.id !== group.id) {
          flushGroup()
        }
        if (!activeGroup || activeGroup.config.id !== group.id) {
          activeGroup = { config: group, steps: [] }
        }
        activeGroup.steps.push(step)
        continue
      }

      if (isComponentActionStep(step.name)) {
        const isPre = /\(pre_deploy_component\)$/.test(step.name)
        const adjacentGroup = findAdjacentComponentGroup(steps, i, isPre ? 'before' : 'after')

        if (adjacentGroup) {
          if (activeGroup && activeGroup.config.id !== adjacentGroup.id) {
            flushGroup()
          }
          if (!activeGroup || activeGroup.config.id !== adjacentGroup.id) {
            activeGroup = { config: { ...adjacentGroup, stepLabels: { ...adjacentGroup.stepLabels } }, steps: [] }
          }
          activeGroup.config.stepLabels![step.name] = getActionStepLabel(step.name)
          activeGroup.steps.push(step)
          continue
        }
      }

      flushGroup()

      items.push({
        id: step.id,
        label: toSentenceCase(step.name),
        status: step.retried ? 'retried' : step.status?.status || 'pending',
        steps: [step],
      })
    }

    flushGroup()
    return items
  }, [steps])
}

function GroupedStepTimeline({ steps, stepLabels }: { steps: TWorkflowStep[]; stepLabels?: Record<string, string> }) {
  return (
    <div className="flex items-start pt-2 pb-4 px-2">
      {steps.map((step, i) => {
        const status = step.retried ? 'retried' : step.status?.status || 'pending'
        const theme = getStatusTheme(status)
        const iconVariant = getStatusIconVariant(status)
        const isLast = i === steps.length - 1
        const isComplete = theme === 'success'
        const nextStep = steps[i + 1]
        const nextComplete = nextStep
          ? getStatusTheme(nextStep.retried ? 'retried' : nextStep.status?.status || 'pending') === 'success'
          : false
        const label = stepLabels?.[step.name] ?? toSentenceCase(step.name)

        return (
          <Fragment key={step.id}>
            <div className="flex flex-col items-center gap-1.5 min-w-0">
              <span
                className={cn(
                  'flex items-center justify-center rounded-full w-7 h-7 shrink-0',
                  TIMELINE_NODE_CLASSES[theme],
                )}
              >
                <Icon variant={iconVariant} weight="bold" size={14} />
              </span>
              <Text variant="label" className="text-center whitespace-nowrap">
                {label}
              </Text>
            </div>
            {!isLast && (
              <div className="flex-1 min-w-6 h-0.5 rounded-full bg-cool-grey-200 dark:bg-cool-grey-700 mt-3.5 mx-1 overflow-hidden">
                <div
                  className="h-full rounded-full bg-green-600 dark:bg-green-500"
                  style={{
                    width: isComplete && nextComplete ? '100%' : isComplete ? '50%' : '0%',
                    transition: 'width 800ms cubic-bezier(0.65, 0, 0.35, 1)',
                  }}
                />
              </div>
            )}
          </Fragment>
        )
      })}
    </div>
  )
}

function OnboardingRunnerDetails({ installId, orgId, shouldPoll }: { installId: string; orgId: string; shouldPoll: boolean }) {
  const { data: install } = useQuery({
    queryKey: ['onboarding-install', installId],
    queryFn: () => getInstall({ installId, orgId }),
    enabled: !!installId && !!orgId,
  })

  const runnerId = install?.runner_id

  const { data: runner } = useQuery({
    queryKey: ['onboarding-runner', runnerId],
    queryFn: () => getRunner({ runnerId: runnerId!, orgId }),
    enabled: !!runnerId,
  })

  const { data: heartbeats } = useQuery({
    queryKey: ['onboarding-runner-heartbeat', runnerId],
    queryFn: () => getRunnerLatestHeartbeat({ runnerId: runnerId!, orgId }),
    enabled: !!runnerId,
    refetchInterval: shouldPoll ? 5000 : false,
  })

  const runnerHeartbeat =
    heartbeats?.install ??
    heartbeats?.org ??
    heartbeats?.build ??
    heartbeats?.[''] ??
    undefined

  if (!runner) {
    return (
      <div className="grid gap-6 md:grid-cols-2 pt-2">
        <LabeledValue label="Status"><Skeleton height="23px" width="75px" /></LabeledValue>
        <LabeledValue label="Connectivity"><Skeleton height="23px" width="110px" /></LabeledValue>
      </div>
    )
  }

  return (
    <div className="grid gap-6 md:grid-cols-2 pt-2">
      <LabeledValue label="Status">
        <Status
          status={runner.status === 'active' ? 'healthy' : 'unhealthy'}
          variant="badge"
        />
      </LabeledValue>
      <LabeledValue label="Connectivity">
        <Status
          status={
            isLessThan30SecondsOld(runnerHeartbeat?.created_at)
              ? 'connected'
              : 'not-connected'
          }
          variant="badge"
        />
      </LabeledValue>
      <LabeledValue label="Version">
        {runnerHeartbeat?.version
          ? <Badge size="sm" variant="code">{runnerHeartbeat.version}</Badge>
          : <Text variant="subtext">Waiting on version</Text>
        }
      </LabeledValue>
      <LabeledValue label="Platform">
        <CloudPlatform
          variant="subtext"
          platform={(runner?.runner_group?.platform || runner?.runner_group?.metadata?.['runner.platform'] || 'unknown') as TCloudPlatform}
          colorVariant="color"
        />
      </LabeledValue>
      <LabeledValue label="Started at">
        <Time variant="subtext" time={runnerHeartbeat?.started_at} />
      </LabeledValue>
      <LabeledValue label="Runner ID">
        <ID theme="default">{runner.id}</ID>
      </LabeledValue>
    </div>
  )
}

const RUNNER_HEALTH_STEP = 'await runner health'

function StepCard({ item, onboarding }: { item: IDisplayItem; onboarding?: TOnboarding }) {
  const isGrouped = item.steps.length > 1

  if (isGrouped) {
    const badgeConfig = getStepBadge(item.steps[item.steps.length - 1])
    return (
      <Card className="px-4 py-3 flex flex-col gap-2">
        <div className="flex items-center gap-3">
          <Status
            isWithoutText
            status={item.status}
            variant="timeline"
            iconSize={16}
          />
          <Text weight="strong" className="flex-1 truncate">
            {item.label}
          </Text>
          {badgeConfig?.children && <Badge {...badgeConfig} size="sm" />}
        </div>
        <GroupedStepTimeline steps={item.steps} stepLabels={item.group?.stepLabels} />
      </Card>
    )
  }

  const badgeConfig = getStepBadge(item.steps[0])
  const isRunnerHealth = item.steps[0]?.name === RUNNER_HEALTH_STEP
  const showRunnerDetails = isRunnerHealth && onboarding?.install_id && onboarding?.org_id

  return (
    <Card className={cn('px-4 py-3', showRunnerDetails ? 'flex flex-col gap-2' : 'flex items-center gap-3')}>
      <div className="flex items-center gap-3">
        <Status
          isWithoutText
          status={item.status}
          variant="timeline"
          iconSize={16}
        />
        <Text weight="strong" className="flex-1 truncate">
          {item.label}
        </Text>
        {badgeConfig?.children && <Badge {...badgeConfig} size="sm" />}
      </div>
      {showRunnerDetails && (
        <OnboardingRunnerDetails
          installId={onboarding.install_id!}
          orgId={onboarding.org_id!}
          shouldPoll={item.status !== 'success'}
        />
      )}
    </Card>
  )
}

export const ProvisioningStep = ({
  onAdvance,
  onGoBack,
  sharedData,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id

  const { workflow, steps, workflowId } = useOnboardingWorkflow(onboarding, setSharedData)

  const isFinished = !!workflow?.finished
  const isError = workflow?.status?.status === 'error'

  const { mutate: completeDeploy, isPending: deployPending } = useMutation({
    mutationFn: () => completeDeployStep({ orgId: orgId! }),
    onSuccess: (ob) => {
      setSharedData('onboarding', ob)
      onAdvance()
    },
  })

  const displayItems = useDisplayItems(steps)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          {isFinished && !isError ? (
            <Icon variant="CheckCircle" size={24} weight="fill" theme="success" />
          ) : isError ? (
            <Icon variant="XCircle" size={24} weight="fill" theme="error" />
          ) : (
            <Icon variant="Loading" size={24} />
          )}
          <Text variant="h3" weight="strong">
            {isFinished && !isError
              ? 'Your app is live'
              : isError
                ? 'Provisioning failed'
                : 'Provisioning...'}
          </Text>
        </div>

        {isError && workflow?.status?.status_human_description && (
          <Banner theme="error">{workflow.status.status_human_description}</Banner>
        )}
      </div>

      <div className="flex flex-col gap-3">
        {!workflowId && (
          <>
            <Skeleton height="52px" width="100%" />
            <Skeleton height="52px" width="100%" />
          </>
        )}

        {displayItems.map((item) => (
          <StepCard key={item.id} item={item} onboarding={onboarding} />
        ))}
      </div>

      <div className="flex justify-between">
        {onGoBack ? (
          <Button type="button" variant="secondary" onClick={onGoBack}>
            <Icon variant="CaretLeft" weight="bold" /> Back
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
          {deployPending ? 'Completing...' : (nextStepTitle ?? 'Continue')} {!deployPending && <Icon variant="CaretRight" weight="bold" />}
        </Button>
      </div>
    </div>
  )
}
