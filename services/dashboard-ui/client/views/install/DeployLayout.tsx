import { Outlet, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { DeployHeader } from '@/components/deploys/DeployHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TabNav } from '@/components/navigation/TabNav'
import { DeployProvider } from '@/providers/deploy-provider'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { getComponent, getWorkflow } from '@/lib'
import type { TComponentType } from '@/types'
import type { TNavLink } from '@/types'

function getTabsForComponentType(
  type?: TComponentType,
  traceEnabled?: boolean
): TNavLink[] {
  const tabs: TNavLink[] = [{ path: '/', text: 'Logs' }]
  if (traceEnabled) {
    tabs.push({ path: '/trace', text: 'Trace' })
  }

  switch (type) {
    case 'terraform_module':
      tabs.push(
        { path: '/plan', text: 'Plan' },
        { path: '/variables', text: 'Variables' },
        { path: '/state', text: 'State' },
        { path: '/outputs', text: 'Outputs' },
      )
      break
    case 'pulumi':
      tabs.push(
        { path: '/plan', text: 'Plan' },
        { path: '/state', text: 'State' },
      )
      break
    case 'helm_chart':
      tabs.push(
        { path: '/plan', text: 'Plan' },
        { path: '/values', text: 'Values' },
        { path: '/outputs', text: 'Outputs' },
      )
      break
    case 'kubernetes_manifest':
      tabs.push(
        { path: '/plan', text: 'Plan' },
        { path: '/manifest', text: 'Manifest' },
      )
      break
    case 'docker_build':
    case 'external_image':
    case 'job':
      tabs.push({ path: '/artifact', text: 'Artifact' })
      break
  }

  return tabs
}

const DeployLayoutInner = () => {
  const { componentId, deployId, installId } = useParams()
  const { deploy } = useDeploy()
  const { install } = useInstall()
  const { org } = useOrg()

  const { data: component } = useQuery({
    queryKey: ['component', org?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, componentId: componentId! }),
    enabled: !!org?.id && !!componentId,
  })

  const { data: workflow } = useQuery({
    queryKey: ['workflow', org?.id, deploy?.install_workflow_id],
    queryFn: () => getWorkflow({ orgId: org.id, workflowId: deploy!.install_workflow_id }),
    enabled: !!org?.id && !!deploy?.install_workflow_id,
  })

  const { hasResponded } = useRespondedApprovals()

  if (!deploy || !component) return null

  const step = workflow?.steps
    ?.filter(
      (s) => s?.step_target_id === deploy?.id && s?.execution_type === 'approval'
    )
    ?.at(-1) ?? null
  const responded = step ? hasResponded(step.id) : false
  const stepStatus = step?.status?.status
  const isTerminal = stepStatus === 'error' || stepStatus === 'cancelled' || stepStatus === 'discarded'
  const isAutoApprove =
    step?.approval?.type === 'approve-all' ||
    step?.approval?.response?.type === 'auto-approve'
  const pendingApproval =
    step?.approval && !step?.approval?.response && !responded && !isTerminal && stepStatus !== 'auto-skipped'

  const basePath = `/${org?.id}/installs/${installId}/components/${componentId}/deploys/${deployId}`
  const tabs = getTabsForComponentType(
    component?.type,
    org?.features?.['trace-view']
  )

  if (pendingApproval && !isAutoApprove) {
    const planTab = tabs.find((t) => t.path === '/plan')
    if (planTab) planTab.badge = true
  }

  return (
    <PageSection>
      <PageTitle title={`Deploy | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: `/${org?.id}/installs/${install?.id}/components`, text: 'Components' },
          {
            path: `/${org?.id}/installs/${install?.id}/components/${componentId}`,
            text: deploy?.component_name,
          },
          {
            path: basePath,
            text: 'Deploy',
          },
        ]}
      />

      <DeployHeader component={component} workflow={workflow} stepId={step?.id} />

      {pendingApproval && !isAutoApprove ? (
        <ApprovalBanner step={step} />
      ) : null}

      <TabNav basePath={basePath} tabs={tabs} />
      <Outlet context={{ component, workflow, step }} />
    </PageSection>
  )
}

export const DeployLayout = () => {
  const { deployId, installId } = useParams()

  return (
    <DeployProvider deployId={deployId!} installId={installId!} shouldPoll>
      <DeployLayoutInner />
    </DeployProvider>
  )
}
