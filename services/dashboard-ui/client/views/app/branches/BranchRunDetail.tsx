import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { WorkflowStepsPipeline } from '@/components/branches/WorkflowStepsPipeline'
import { WorkflowStepDetail } from '@/components/branches/WorkflowStepDetail'
import { AppConfigDiff } from '@/components/branches/AppConfigDiff'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { BranchProvider } from '@/providers/branch-provider'
import { isActiveStepStatus } from '@/components/branches/shared/step-status'
import { ConfigDiffFocusContext, type TConfigDiffFocus } from '@/components/approvals/plan-diffs/config-diff-focus'
import { getBranchWorkflowRun } from '@/lib'
import { toSentenceCase } from '@/utils/string-utils'
import { scrollElementIntoView } from '@/utils/scroll'
import { useSearchParamState } from '@/hooks/use-search-param-state'
import { useCallback, useEffect, useRef, useState } from 'react'

const WORKFLOW_TYPE_LABELS: Record<string, string> = {
  app_branches_manual_update: 'Manual update',
  app_branches_config_repo_update: 'Config update',
  app_branches_component_repo_update: 'Component update',
  app_branch_config_update: 'Config update',
}

const BranchRunDetailContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string
  const runId = params.runId as string
  const [urlStepId, setUrlStepId] = useSearchParamState('step')
  const stepDetailRef = useRef<HTMLDivElement>(null)
  const pendingScrollRef = useRef(false)
  const [configFocus, setConfigFocus] = useState<TConfigDiffFocus | null>(null)
  const requestConfigFocus = useCallback((sectionKey: string, entityName?: string) => {
    setConfigFocus((prev) => ({ sectionKey, entityName, nonce: (prev?.nonce ?? 0) + 1 }))
  }, [])

  const { data: run, isLoading } = useQuery({
    queryKey: ['branch-run', orgId, appId, branchId, runId],
    queryFn: () => getBranchWorkflowRun({ orgId, appId, branchId, runId }),
    enabled: !!orgId && !!appId && !!branchId && !!runId,
    refetchInterval: 5000,
  })

  const steps = (run?.steps || []).filter((s) => s.owner_type !== 'components')
  const activeStep = steps.find((step) => isActiveStepStatus(step.status?.status))

  const urlStep = urlStepId ? steps.find((s) => s.id === urlStepId) ?? null : null
  const selectedStep = urlStep ?? activeStep ?? steps[0] ?? null
  const selectedStepId = selectedStep?.id ?? null

  useEffect(() => {
    if (pendingScrollRef.current) {
      scrollElementIntoView(stepDetailRef.current, { block: 'start' })
      pendingScrollRef.current = false
    }
  }, [selectedStepId])

  const handleJumpToActive = () => {
    if (!activeStep) return
    if (selectedStepId === activeStep.id) {
      scrollElementIntoView(stepDetailRef.current, { block: 'start' })
      return
    }
    pendingScrollRef.current = true
    setUrlStepId(activeStep.id ?? null)
  }

  const configStep = steps.find((s) => s.name?.toLowerCase().includes('config') && !s.name?.toLowerCase().includes('diff'))
  const appConfigId = configStep?.status?.metadata?.app_config_id as string | undefined

  if (isLoading || !run) {
    return (
      <PageSection>
        <Text variant="body" theme="neutral">
          Loading workflow run...
        </Text>
      </PageSection>
    )
  }

  const status = run.status?.status || 'unknown'
  const statusDescription = run.status?.status_human_description || ''

  const branchRun = run.app_branch_runs?.at(0)
  const commitMessage = branchRun?.vcs_connection_commit?.message?.split('\n')[0]?.trim()
  const typeLabel = run.type ? WORKFLOW_TYPE_LABELS[run.type] : undefined
  const runTitle = toSentenceCase(commitMessage || typeLabel || run.name || 'Workflow run')

  return (
    <ConfigDiffFocusContext.Provider value={{ requestFocus: requestConfigFocus }}>
    <PageSection className="max-w-full">
      <PageTitle title={`${runTitle} | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}`, text: branch?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}/runs/${runId}`, text: 'Run' },
        ]}
      />

      <BackLink />

      <div className="flex items-start justify-between gap-4">
        <HeadingGroup className="gap-1.5 min-w-0">
          <div className="flex items-center gap-2.5">
            <Text as="h1" variant="h2" weight="strong" className="leading-tight min-w-0 truncate" title={runTitle}>
              {runTitle}
            </Text>
            {branch?.name && (
              <Badge size="sm" variant="code" className="shrink-0">
                {branch.name}
              </Badge>
            )}
          </div>

          <ID className="text-[12px] font-mono text-cool-grey-400 dark:text-cool-grey-500">{runId}</ID>

          <div className="flex items-center gap-2 mt-0.5">
            <Status status={status} variant="badge" />
            {statusDescription && (
              <Text variant="subtext" theme="neutral">{statusDescription}</Text>
            )}
          </div>
        </HeadingGroup>

        <div className="flex flex-col items-end gap-2 shrink-0">
          <div className="flex flex-col items-end gap-0.5">
            <div className="flex items-center gap-1.5">
              <Text variant="subtext" theme="neutral">Created</Text>
              <Time time={run.created_at} format="relative" variant="subtext" />
            </div>
            {run.started_at && (
              <div className="flex items-center gap-1.5">
                <Text variant="subtext" theme="neutral">Started</Text>
                <Time time={run.started_at} format="relative" variant="subtext" />
              </div>
            )}
            {run.finished_at && (
              <div className="flex items-center gap-1.5">
                <Text variant="subtext" theme="neutral">Finished</Text>
                <Time time={run.finished_at} format="relative" variant="subtext" />
              </div>
            )}
          </div>

          <div className="flex items-center gap-2">
            <AdminDashboardLink path={`/workflows/${runId}`} label="admin" />
            <CancelWorkflowButton workflow={run} />
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-3">
          <Text variant="h3" weight="strong">
            Workflow progress
          </Text>
          {activeStep && (
            <Button variant="secondary" onClick={handleJumpToActive}>
              Jump to active step
            </Button>
          )}
        </div>
        <WorkflowStepsPipeline
          steps={steps}
          selectedStepId={selectedStep?.id}
          onSelectStep={(step) => setUrlStepId(step?.id ?? null)}
        />
      </div>

      {appConfigId && <AppConfigDiff appConfigId={appConfigId} focus={configFocus} />}

      {selectedStep && (
        <div className="flex flex-col gap-2">
          <div ref={stepDetailRef} className="flex items-baseline gap-3 scroll-mt-4">
            <Text variant="h3" weight="strong">Step details</Text>
            <Text variant="subtext" theme="neutral">{selectedStep.name}</Text>
          </div>
          <WorkflowStepDetail
            step={selectedStep}
            appBranchRunId={branchRun?.id}
            onClose={() => setUrlStepId(null)}
          />
        </div>
      )}
    </PageSection>
    </ConfigDiffFocusContext.Provider>
  )
}

export const BranchRunDetail = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId}>
      <BranchRunDetailContent />
    </BranchProvider>
  )
}
