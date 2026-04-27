import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Plan } from '@/components/approvals/Plan'
import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { SandboxHeader } from '@/components/sandbox/SandboxHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import { SandboxRunProvider } from '@/providers/sandbox-run-provider'
import { useSandboxRun } from '@/hooks/use-sandbox-run'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { getWorkflow } from '@/lib'

export const SandboxRunDetail = () => {
  const { runId } = useParams()

  return (
    <SandboxRunProvider runId={runId!} shouldPoll>
      <SandboxRunDetailContent />
    </SandboxRunProvider>
  )
}

const SandboxRunDetailContent = () => {
  const { runId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const { sandboxRun } = useSandboxRun()

  const { data: workflow } = useQuery({
    queryKey: ['workflow', org?.id, sandboxRun?.install_workflow_id],
    queryFn: () =>
      getWorkflow({ orgId: org.id, workflowId: sandboxRun!.install_workflow_id }),
    enabled: !!org?.id && !!sandboxRun?.install_workflow_id,
  })

  const step = workflow?.steps
    ?.filter((s) => s?.step_target_id === sandboxRun?.id)
    ?.at(-1) ?? null

  const { hasResponded } = useRespondedApprovals()
  const responded = step ? hasResponded(step.id) : false
  const logStream = sandboxRun?.log_stream
  const stepStatus = step?.status?.status
  const isTerminal = stepStatus === 'error' || stepStatus === 'cancelled' || stepStatus === 'discarded'
  const isAutoApprove =
    step?.approval?.type === 'approve-all' ||
    step?.approval?.response?.type === 'auto-approve'
  const pendingApproval =
    step?.approval && !step?.approval?.response && !responded && !isTerminal && stepStatus !== 'auto-skipped'
  const completedApproval =
    step?.approval && (!!step?.approval?.response || responded) && !isTerminal && stepStatus !== 'auto-skipped'
  const showPlanBelow = completedApproval || isAutoApprove

  return (
    <PageSection flush>
      <PageTitle title={`Sandbox run | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: `/${org?.id}/installs/${install?.id}/sandbox`, text: 'Sandbox' },
          {
            path: `/${org?.id}/installs/${install?.id}/sandbox/runs/${runId}`,
            text: sandboxRun?.run_type ?? 'Run',
          },
        ]}
      />

      <SandboxHeader workflow={workflow} stepId={step?.id} />

      <PageSection className="!pb-12">
        <div className="flex flex-col gap-6">
          {pendingApproval && !isAutoApprove ? (
            <div className="flex flex-col gap-4">
              <ApprovalBanner step={step} />
              <Plan step={step} />
            </div>
          ) : null}

          {logStream ? (
            <LogStreamProvider logStreamId={logStream.id} shouldPoll={logStream.open}>
              <UnifiedLogsProvider>
                <LogViewerProvider>
                  <SSELogs />
                </LogViewerProvider>
              </UnifiedLogsProvider>
            </LogStreamProvider>
          ) : (
            <LogsSkeleton />
          )}

          {showPlanBelow && step ? (
            <div className="flex flex-col gap-4">
              {!isAutoApprove && <ApprovalBanner step={step} />}
              <Plan step={step} />
            </div>
          ) : null}
        </div>

      </PageSection>
    </PageSection>
  )
}
