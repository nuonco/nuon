import { Outlet, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { SandboxHeader } from '@/components/sandbox/SandboxHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TabNav } from '@/components/navigation/TabNav'
import { SandboxRunProvider } from '@/providers/sandbox-run-provider'
import { useSandboxRun } from '@/hooks/use-sandbox-run'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { getWorkflow } from '@/lib'
import type { TNavLink } from '@/types'

const sandboxTabs: TNavLink[] = [
  { path: '/', text: 'Logs' },
  { path: '/trace', text: 'Trace' },
  { path: '/plan', text: 'Plan' },
  { path: '/variables', text: 'Variables' },
  { path: '/state', text: 'State' },
  { path: '/outputs', text: 'Outputs' },
]

const SandboxRunLayoutInner = () => {
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
    ?.filter(
      (s) => s?.step_target_id === sandboxRun?.id && s?.execution_type === 'approval'
    )
    ?.at(-1) ?? null

  const { hasResponded } = useRespondedApprovals()
  const responded = step ? hasResponded(step.id) : false
  const stepStatus = step?.status?.status
  const isTerminal = stepStatus === 'error' || stepStatus === 'cancelled' || stepStatus === 'discarded'
  const isAutoApprove =
    step?.approval?.type === 'approve-all' ||
    step?.approval?.response?.type === 'auto-approve'
  const pendingApproval =
    step?.approval && !step?.approval?.response && !responded && !isTerminal && stepStatus !== 'auto-skipped'

  const basePath = `/${org?.id}/installs/${install?.id}/sandbox/runs/${runId}`
  const traceEnabled = !!org?.features?.['trace-view']
  const tabs = sandboxTabs
    .filter((t) => traceEnabled || t.path !== '/trace')
    .map((t) => ({ ...t }))

  if (pendingApproval && !isAutoApprove) {
    const planTab = tabs.find((t) => t.path === '/plan')
    if (planTab) planTab.badge = true
  }

  return (
    <PageSection>
      <PageTitle title={`Sandbox run | ${install?.name}`} />
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

      <SandboxHeader workflow={workflow} stepId={step?.id} flush />

      {pendingApproval && !isAutoApprove ? (
        <ApprovalBanner step={step} />
      ) : null}

      <TabNav basePath={basePath} tabs={tabs} />
      <Outlet context={{ workflow, step }} />
    </PageSection>
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
