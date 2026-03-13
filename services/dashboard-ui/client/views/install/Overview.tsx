import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { InstallTopology } from '@/components/installs/overview/InstallTopology'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallWorkflows, getWorkflow } from '@/lib'
import type { TCloudPlatform, TWorkflowStep } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

function MetadataBar() {
  const { install } = useInstall()

  const driftCount = install?.drifted_objects?.length ?? 0
  const cloudPlatform = (install?.cloud_platform ?? 'unknown') as TCloudPlatform
  const region =
    install?.aws_account?.region ??
    install?.azure_account?.location ??
    ''
  const runnerStatus = install?.runner_status ?? ''
  const stackStatus =
    install?.install_stack?.versions?.at(-1)?.composite_status?.status ??
    runnerStatus

  const sep = <div className="w-px h-3.5 bg-cool-grey-200 dark:bg-dark-grey-600 shrink-0" />

  return (
    <div className="flex items-center gap-3 px-4 py-3 border-b">
      {/* Cloud platform + Region */}
      {region && (
        <>
          <div className="flex items-center gap-1.5">
            <CloudPlatform platform={cloudPlatform} displayVariant="icon-only" iconSize="16" theme="neutral" />
            <Text variant="body" theme="neutral">{region}</Text>
          </div>
          {sep}
        </>
      )}

      {/* Stack status + version */}
      {stackStatus && (
        <>
          <Link href="stacks" className="no-underline">
            <div className="flex items-center gap-1.5">
              <Icon variant="SneakerMove" size={16} weight="bold" className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
              <Status status={stackStatus} variant="badge" />
            </div>
          </Link>
          {sep}
        </>
      )}

      {/* Drift */}
      <Link href="workflows" className="no-underline ml-auto">
        <div className="flex items-center gap-1.5">
          <Icon variant="Scan" size={16} weight="bold" className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
          {driftCount > 0 ? (
            <Text variant="body" theme="warn">
              {driftCount} {driftCount === 1 ? 'Drift' : 'Drifts'} Detected
            </Text>
          ) : (
            <Text variant="body" theme="neutral">No Drifts Detected</Text>
          )}
        </div>
      </Link>
    </div>
  )
}

const RUNNING_STEP_STATUSES = new Set([
  'in-progress','provisioning','building','applying','planning',
  'checking-plan','generating','retrying','deleting','executing',
  'started','syncing','deploying',
])

function isStepRunning(step: TWorkflowStep): boolean {
  return RUNNING_STEP_STATUSES.has(step.status?.status ?? '')
}

function useElapsed(startedAt: string | undefined): string {
  const [, tick] = useState(0)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    if (!startedAt) return
    intervalRef.current = setInterval(() => tick((n) => n + 1), 1000)
    return () => { if (intervalRef.current) clearInterval(intervalRef.current) }
  }, [startedAt])

  if (!startedAt) return ''
  const ms = Date.now() - new Date(startedAt).getTime()
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const h = Math.floor(m / 60)
  if (h > 0) return `${h}h ${m % 60}m`
  if (m > 0) return `${m}m ${s % 60}s`
  return `${s}s`
}

function ActiveWorkflowBar() {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: workflowsResult } = useQuery({
    queryKey: ['install-active-workflow-v2', org?.id, install?.id],
    queryFn: () =>
      getInstallWorkflows({
        installId: install.id,
        orgId: org.id,
        limit: 10,
        offset: 0,
        planonly: false,
      }),
    enabled: !!org?.id && !!install?.id,
    refetchInterval: 10_000,
    refetchOnMount: 'always',
    staleTime: 0,
  })

  const workflows = workflowsResult?.data ?? []
  const activeWorkflow = workflows.find((w) => w.finished !== true)
  const workflow = activeWorkflow ?? workflows[0]
  const hasActiveWorkflow = !!activeWorkflow

  // Secondary fetch: get live step data from the workflow detail endpoint
  // when there's an active workflow — the list endpoint may not always
  // include fully up-to-date steps.
  const { data: liveWorkflow } = useQuery({
    queryKey: ['install-active-workflow-detail', org?.id, workflow?.id],
    queryFn: () => getWorkflow({ workflowId: workflow!.id!, orgId: org.id }),
    enabled: !!org?.id && !!workflow?.id,
    refetchInterval: hasActiveWorkflow ? 5_000 : false,
    refetchOnMount: 'always',
    staleTime: 0,
  })

  // Use live step data when available, fall back to list data
  const resolvedWorkflow = liveWorkflow ?? workflow

  // Visible steps (exclude hidden execution_type)
  const allSteps = resolvedWorkflow?.steps ?? []
  const steps = allSteps.filter((s) => s.execution_type !== 'hidden')

  const needsApprovalStep = steps.find((s) => s.status?.status === 'approval-awaiting')
  const activeStep = steps.find((s) => s.started_at && !s.finished && !needsApprovalStep) ??
    steps.find((s) => isStepRunning(s))

  const workflowLabel =
    resolvedWorkflow?.name ??
    toSentenceCase((resolvedWorkflow?.type ?? '').replace(/[-_]/g, ' '))

  const elapsed = useElapsed(hasActiveWorkflow ? resolvedWorkflow?.started_at : undefined)

  const lastRunFailed = !hasActiveWorkflow && resolvedWorkflow?.status?.status === 'error'
  const lastRunFinishedAt = !hasActiveWorkflow && resolvedWorkflow?.finished_at
    ? new Date(workflow.finished_at).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
    : null

  const dot = <span className="text-cool-grey-300 dark:text-dark-grey-600 select-none px-0.5">·</span>

  return (
    <div className="flex items-center gap-2 px-4 py-2 border-t min-h-[40px] overflow-hidden">

      {/* LEFT — identity */}
      <div className="flex items-center gap-2 shrink-0">
        {hasActiveWorkflow ? (
          <span className="relative flex h-2 w-2 shrink-0">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75" />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500" />
          </span>
        ) : lastRunFailed ? (
          <span className="h-2 w-2 rounded-full bg-red-500 shrink-0" />
        ) : (
          <span className="h-2 w-2 rounded-full bg-cool-grey-300 dark:bg-dark-grey-500 shrink-0" />
        )}

        {workflow ? (
          <Link href={`workflows/${workflow.id}`} className="no-underline shrink-0">
            <Text variant="body" className={hasActiveWorkflow ? 'text-primary-500 dark:text-primary-400' : lastRunFailed ? 'text-red-600 dark:text-red-400' : 'text-cool-grey-600 dark:text-cool-grey-400'}>
              {workflowLabel || 'Workflow'}
            </Text>
          </Link>
        ) : (
          <Text variant="body" theme="neutral">No workflows yet</Text>
        )}

        {workflow?.plan_only && (
          <span className="text-[10px] font-medium px-1.5 py-0.5 rounded bg-cool-grey-100 dark:bg-dark-grey-700 text-cool-grey-500 dark:text-cool-grey-400 border border-cool-grey-200 dark:border-dark-grey-600 shrink-0">
            Drift scan
          </span>
        )}
      </div>

      {/* CENTRE — step info */}
      {steps.length > 0 && (
        <>
          {dot}
          <Text variant="body" theme="neutral" className="shrink-0 text-[12px] tabular-nums">
            {steps.filter((s) => s.finished).length}/{steps.length} steps
          </Text>
          {activeStep?.name && (
            <>
              {dot}
              <Text variant="body" theme="neutral" className="truncate text-[12px] max-w-[240px]">
                {toSentenceCase((activeStep.name).replace(/[-_]/g, ' '))}
              </Text>
            </>
          )}
        </>
      )}

      {/* RIGHT — approval callout / elapsed / idle state */}
      <div className="ml-auto flex items-center gap-2 shrink-0">
        {needsApprovalStep ? (
          <Link href={`workflows/${workflow?.id}`} className="no-underline">
            <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-orange-50 dark:bg-orange-950 border border-orange-200 dark:border-orange-800">
              <Icon variant="Warning" size={12} weight="fill" className="text-orange-500 shrink-0" />
              <Text variant="body" className="text-orange-600 dark:text-orange-400 text-[12px]">Approval required</Text>
            </div>
          </Link>
        ) : hasActiveWorkflow && elapsed ? (
          <Text variant="body" theme="neutral" className="text-[12px]">{elapsed}</Text>
        ) : lastRunFailed ? (
          <Link href={`workflows/${workflow?.id}`} className="no-underline">
            <Text variant="body" className="text-red-600 dark:text-red-400 text-[12px]">View failure</Text>
          </Link>
        ) : lastRunFinishedAt ? (
          <Link href={`workflows/${workflow?.id}`} className="no-underline">
            <Text variant="body" theme="neutral" className="text-[12px]">Last run {lastRunFinishedAt}</Text>
          </Link>
        ) : null}
      </div>
    </div>
  )
}

export const Overview = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection className="!pt-0" isScrollable>
      <PageTitle title={`Overview | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
        ]}
      />

      <div className="p-6">
        <div className="w-full border rounded-lg overflow-hidden">
          <MetadataBar />
          <InstallTopology />
          <ActiveWorkflowBar />
        </div>
      </div>
    </PageSection>
  )
}
