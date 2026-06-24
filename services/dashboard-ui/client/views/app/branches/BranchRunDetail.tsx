import { useParams, useSearchParams } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toast } from '@/components/surfaces/Toast'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { WorkflowStepsPipeline } from '@/components/branches/WorkflowStepsPipeline'
import { WorkflowStepDetail } from '@/components/branches/WorkflowStepDetail'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useToast } from '@/hooks/use-toast'
import { BranchProvider } from '@/providers/branch-provider'
import { getBranchWorkflowRun, cancelWorkflow } from '@/lib'
import { useEffect, useState } from 'react'
import type { TAPIError } from '@/types'

function statusTheme(status?: string) {
  if (status === 'success') return 'success'
  if (status === 'error') return 'error'
  if (status === 'in-progress') return 'info'
  return 'neutral'
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
  const [searchParams] = useSearchParams()
  const targetStepId = searchParams.get('target')
  const [selectedStepId, setSelectedStepId] = useState<string | null>(targetStepId)

  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { data: run, isLoading } = useQuery({
    queryKey: ['branch-run', orgId, appId, branchId, runId],
    queryFn: () => getBranchWorkflowRun({ orgId, appId, branchId, runId }),
    enabled: !!orgId && !!appId && !!branchId && !!runId,
    refetchInterval: 5000,
  })

  const { mutate: cancel, isPending: isCancelling } = useMutation({
    mutationFn: () => cancelWorkflow({ orgId, workflowId: runId }),
    onSuccess: () => {
      addToast(
        <Toast heading="Workflow cancelled" theme="success">
          <Text>The workflow run has been cancelled.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['branch-run', orgId, appId, branchId, runId] })
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Cancel failed" theme="error">
          <Text>{err?.error || 'Unable to cancel workflow.'}</Text>
        </Toast>
      )
    },
  })

  const steps = (run?.steps || []).filter((s) => s.owner_type !== 'components')

  useEffect(() => {
    if (steps.length > 0 && !selectedStepId) {
      const inProgressStep = steps.find((step) => step.status?.status === 'in-progress')
      setSelectedStepId((inProgressStep || steps[0])?.id ?? null)
    }
  }, [steps, selectedStepId])

  const selectedStep = selectedStepId ? steps.find((s) => s.id === selectedStepId) ?? null : null

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
  const isActive = ['pending', 'queued', 'in-progress', 'approval-awaiting'].includes(status)

  return (
    <PageSection className="max-w-full space-y-4">
      <PageTitle title={`Run | ${app?.name}`} />
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

      {/* ── Page header ── */}
      <div className="flex items-start justify-between gap-4">
        {/* Left: title + run id + status */}
        <div className="flex flex-col gap-1.5 min-w-0">
          <div className="flex items-center gap-2.5">
            <h1 className="text-[22px] font-semibold text-cool-grey-900 dark:text-white leading-tight">
              Workflow run
            </h1>
            {branch?.name && (
              <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full border border-cool-grey-300 dark:border-dark-grey-600 bg-cool-grey-50 dark:bg-dark-grey-800 font-mono text-[12px] text-cool-grey-600 dark:text-cool-grey-300 shrink-0">
                <svg width="12" height="12" viewBox="0 0 16 16" fill="none" className="text-cool-grey-400 dark:text-cool-grey-500">
                  <path d="M5 3a2 2 0 1 0 0 4 2 2 0 0 0 0-4zm0 6a2 2 0 1 0 0 4 2 2 0 0 0 0-4zm6-6a2 2 0 1 0 0 4 2 2 0 0 0 0-4z" fill="currentColor" fillOpacity=".6" />
                  <path d="M5 7v2M5 9a4 4 0 0 0 4 4h2" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
                </svg>
                {branch.name}
              </span>
            )}
          </div>

          <div className="flex items-center gap-1.5">
            <ID className="text-[12px] font-mono text-cool-grey-400 dark:text-cool-grey-500">{runId}</ID>
          </div>

          <div className="flex items-center gap-2 mt-0.5">
            <Badge theme={statusTheme(status)} size="sm">
              {status === 'in-progress' && (
                <svg className="animate-spin w-3 h-3 shrink-0" viewBox="0 0 12 12" fill="none">
                  <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeOpacity="0.3" strokeWidth="1.5" />
                  <path d="M6 1.5 A4.5 4.5 0 0 1 10.5 6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
                </svg>
              )}
              {status}
            </Badge>
            {statusDescription && (
              <Text variant="subtext" theme="neutral">{statusDescription}</Text>
            )}
          </div>
        </div>

        {/* Right: timestamps + actions */}
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
            {isActive && (
              <Button
                variant="danger"
                size="sm"
                onClick={() => cancel()}
                disabled={isCancelling}
              >
                <Icon variant="XCircleIcon" size={15} />
                {isCancelling ? 'Cancelling...' : 'Cancel run'}
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* ── Workflow progress card ── */}
      <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm">
        <div className="flex items-center justify-between px-5 py-4 border-b border-cool-grey-100 dark:border-dark-grey-800">
          <Text variant="h3" weight="strong">
            Workflow progress
          </Text>
          <Text variant="subtext" theme="neutral" className="cursor-pointer hover:underline">
            Jump to a step
          </Text>
        </div>
        <div className="px-4 pb-4">
          <WorkflowStepsPipeline
            steps={steps}
            selectedStepId={selectedStep?.id}
            onSelectStep={(step) => setSelectedStepId(step?.id ?? null)}
          />
        </div>
      </div>

      {/* ── Step detail card ── */}
      {selectedStep && (
        <>
          <div className="flex items-baseline gap-3 mt-2">
            <Text variant="h3" weight="strong">Step details</Text>
            <Text variant="subtext" theme="neutral">{selectedStep.name}</Text>
          </div>
          <WorkflowStepDetail
            step={selectedStep}
            onClose={() => setSelectedStepId(null)}
          />
        </>
      )}
    </PageSection>
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
