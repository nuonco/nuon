import { Outlet, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { CompositeError } from '@/components/common/CompositeError'
import { SandboxHeader } from '@/components/sandbox/SandboxHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { SandboxRunProvider } from '@/providers/sandbox-run-provider'
import { useSandboxRun } from '@/hooks/use-sandbox-run'
import { useInstallPage } from '@/hooks/use-install-path'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { getWorkflow } from '@/lib'
import type { TNavLink } from '@/types'

const sandboxTabs: TNavLink[] = [
  { path: '/', text: 'Summary' },
  { path: '/logs', text: 'Logs' },
  { path: '/trace', text: 'Trace' },
  { path: '/plan', text: 'Plan' },
  { path: '/variables', text: 'Variables' },
  { path: '/state', text: 'State' },
  { path: '/outputs', text: 'Outputs' },
]

const SandboxRunLayoutInner = () => {
  const { runId } = useParams()
  const { org, install, href } = useInstallPage()
  const { sandboxRun } = useSandboxRun()

  const { data: workflow } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['workflow', org?.id, sandboxRun?.install_workflow_id],
    queryFn: () =>
      getWorkflow({ orgId: org.id, workflowId: sandboxRun!.install_workflow_id }),
    enabled: !!org?.id && !!sandboxRun?.install_workflow_id,
  })

  const step = workflow?.steps
    ?.filter(
      (s) => s?.step_target_id === sandboxRun?.id && s?.execution_type === 'approval'
    )
    ?.at(-1) ?? null

  const { hasResponded } = useRespondedApprovals()
  const responded = step ? hasResponded(step.id) : false
  const stepStatus = step?.status?.status
  const isTerminal =
    stepStatus === 'error' ||
    stepStatus === 'cancelled' ||
    stepStatus === 'discarded' ||
    stepStatus === 'approval-expired' ||
    stepStatus === 'approval-denied'
  const isAutoApprove =
    step?.approval?.type === 'approve-all' ||
    step?.approval?.response?.type === 'auto-approve'
  const pendingApproval =
    step?.approval && !step?.approval?.response && !responded && !isTerminal && stepStatus !== 'auto-skipped'

  const basePath = href(`/sandbox/runs/${runId}`)
  const tabs = sandboxTabs.map((t) => ({ ...t }))

  if (pendingApproval && !isAutoApprove) {
    const planTab = tabs.find((t) => t.path === '/plan')
    if (planTab) planTab.badge = true
  }

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: `/${org?.id}/installs/${install?.id}/sandbox`, text: 'Sandbox' },
          {
            path: basePath,
            text: sandboxRun?.run_type ?? 'Run',
          },
        ]}
      />

      <DetailPage
        header={<SandboxHeader workflow={workflow} stepId={step?.id} />}
        banners={
          <>
            {sandboxRun?.composite_error ? (
              <CompositeError error={sandboxRun.composite_error} />
            ) : null}
            {pendingApproval && !isAutoApprove ? (
              <ApprovalBanner step={step} />
            ) : null}
          </>
        }
        tabNav={{ basePath, tabs }}
      >
        <Outlet context={{ workflow, step }} />
      </DetailPage>
    </>
  )
}

export const SandboxRunLayout = () => {
  const { installId, runId } = useParams()

  return (
    <SandboxRunProvider installId={installId!} runId={runId!} shouldPoll>
      <SandboxRunLayoutInner />
    </SandboxRunProvider>
  )
}
