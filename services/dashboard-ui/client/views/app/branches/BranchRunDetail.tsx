import { useParams } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
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
import type { TAPIError, TInstallWorkflowStep } from '@/types'

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
  const [selectedStep, setSelectedStep] = useState<TInstallWorkflowStep | null>(null)

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

  // Filter out build sub-steps (owner_type "components") — their status
  // is tracked via the parent builds step's metadata instead.
  const steps = (run?.steps || []).filter((s) => s.owner_type !== 'components')

  useEffect(() => {
    if (steps.length > 0 && !selectedStep) {
      const inProgressStep = steps.find(
        (step) => step.status?.status === 'in-progress'
      )
      setSelectedStep(inProgressStep || steps[0])
    }
  }, [steps, selectedStep])

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
  const branchRun = (run as any)?.app_branch_runs?.[0]
  const commit = branchRun?.vcs_connection_commit

  return (
    <PageSection className="max-w-full">
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
      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="strong">
            Workflow run
          </Text>
          <ID>{runId}</ID>
          <div className="flex items-center gap-3 mt-2">
            <Badge theme={statusTheme(status)} size="sm">
              {status}
            </Badge>
            {statusDescription && (
              <Text variant="subtext" theme="neutral">
                {statusDescription}
              </Text>
            )}
          </div>
        </HeadingGroup>
        <div className="flex flex-col items-end gap-2">
          <div className="flex flex-col items-end gap-1">
            <Text variant="subtext" theme="neutral">
              Created <Time time={run.created_at} format="relative" />
            </Text>
            {run.started_at && (
              <Text variant="subtext" theme="neutral">
                Started <Time time={run.started_at} format="relative" />
              </Text>
            )}
            {run.finished_at && (
              <Text variant="subtext" theme="neutral">
                Finished <Time time={run.finished_at} format="relative" />
              </Text>
            )}
          </div>
          <div className="flex items-center gap-2">
            <AdminDashboardLink path={`/workflows/${runId}`} label="View in admin" />
            {['pending', 'queued', 'in-progress'].includes(status) && (
              <Button
                variant="danger"
                size="sm"
                onClick={() => cancel()}
                disabled={isCancelling}
              >
                <Icon variant="XCircleIcon" size={16} />
                {isCancelling ? 'Cancelling...' : 'Cancel run'}
              </Button>
            )}
          </div>
        </div>
      </div>

      {commit && (
        <div className="flex items-start gap-3 p-4 bg-cool-grey-50 dark:bg-dark-grey-850 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
          <Icon variant="GitCommitIcon" size={20} className="text-cool-grey-500 dark:text-dark-grey-300 mt-0.5 shrink-0" />
          <div className="min-w-0">
            <Text variant="base" weight="strong" className="truncate">
              {commit.message?.split('\n')[0] || 'No message'}
            </Text>
            <div className="flex items-center gap-3 mt-1">
              <Text variant="subtext" theme="neutral" family="mono">
                {commit.sha?.substring(0, 8)}
              </Text>
              {commit.author_name && (
                <Text variant="subtext" theme="neutral">
                  by {commit.author_name}
                </Text>
              )}
              {branchRun?.event_type && (
                <Badge theme="neutral" size="sm">
                  {branchRun.event_type}
                </Badge>
              )}
            </div>
          </div>
        </div>
      )}

      <Card>
        <div className="p-6 min-w-0">
          <div className="flex items-center justify-between mb-4">
            <Text variant="h3" weight="strong">
              Workflow progress
            </Text>
            <Text variant="subtext" theme="neutral">
              Scroll horizontally or use trackpad to navigate
            </Text>
          </div>

          <WorkflowStepsPipeline
            steps={steps}
            selectedStepId={selectedStep?.id}
            onSelectStep={setSelectedStep}
          />
        </div>
      </Card>

      {selectedStep && (
        <WorkflowStepDetail
          step={selectedStep}
          onClose={() => setSelectedStep(null)}
        />
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
