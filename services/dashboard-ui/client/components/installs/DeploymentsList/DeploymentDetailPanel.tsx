import { useState } from 'react'
import { useLocation } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import {
  DeploymentAlerts,
  DeploymentChangesContent,
  DeploymentTemplateContent,
  DeploymentWorkflowContent,
} from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import {
  DeploymentRunStatus,
  ResourceOutcomes,
} from '@/components/installs/DeploymentDetail/DeploymentProgress'
import {
  DEPLOYMENT_TABS,
  deploymentOutcomes,
  deploymentSteps,
  deploymentTabOrder,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { useWorkflow } from '@/hooks/use-workflow'
import { useInstallNested } from '@/hooks/use-install-path'
import { getInstallDeployment } from '@/lib'
import { installHref } from '@/lib/install-path'
import { WorkflowProvider } from '@/providers/workflow-provider'
import type { TInstallDeploymentSummary, TWorkflow } from '@/types'

export interface IDeploymentDetailPanel extends IPanel {
  deployment: TInstallDeploymentSummary
  orgId: string
  appId: string
  installId: string
  repo?: string
}

const DeploymentPanelContent = ({
  deployment,
  orgId,
  appId,
  installId,
  repo,
  workflow,
}: IDeploymentDetailPanel & { workflow?: TWorkflow }) => {
  const { search } = useLocation()
  const nested = useInstallNested()
  const {
    data: record,
    isPending: recordPending,
    isError: recordError,
    refetch: refetchRecord,
  } = useQuery({
    queryKey: ['install-deployment', orgId, installId, deployment.id],
    queryFn: () =>
      getInstallDeployment({ orgId, installId, workflowId: deployment.id }),
    enabled: !!orgId && !!installId,
  })
  const branchHref = record?.app_branch
    ? `/${orgId}/apps/${appId}/branches/${record.app_branch.id}`
    : undefined
  const run = {
    status: workflow?.status?.status ?? deployment.status,
    activity:
      workflow?.status?.status_human_description ?? deployment.activity ?? '',
    steps: deploymentSteps(workflow),
    outcomes: deploymentOutcomes(record, workflow),
  }
  const [tabOrder] = useState(() => deploymentTabOrder(run.status))
  const basePath = installHref({
    orgId,
    appId,
    installId,
    nested,
    suffix: `/deployments/${deployment.id}`,
  })
  const noWorkflow = (
    <EmptyState
      emptyTitle="No workflow"
      emptyMessage="This deployment has no associated workflow."
    />
  )
  const content = {
    template: (
      <DeploymentTemplateContent
        deployment={record}
        appId={appId}
        workflow={workflow}
        isLoading={recordPending}
        isError={recordError}
        onRetry={() => refetchRecord()}
      />
    ),
    workflow: workflow ? <DeploymentWorkflowContent /> : noWorkflow,
    changes: workflow ? (
      <DeploymentChangesContent deployment={record} />
    ) : (
      noWorkflow
    ),
  }
  return (
    <>
      <div className="flex flex-wrap items-center gap-3">
        <DeploymentRunStatus run={run} />
        <ID>{deployment.id}</ID>
        <Time
          time={deployment.created_at}
          format="relative"
          variant="subtext"
        />
      </div>
      <Text
        variant="subtext"
        theme={run.status === 'error' ? 'error' : 'neutral'}
      >
        {run.activity}
      </Text>
      {record?.app_branch && branchHref ? (
        <LabeledValue label="App branch">
          <span className="flex flex-wrap items-center gap-2">
            <Link href={branchHref}>{record.app_branch.name}</Link>
            <CommitLink sha={record.app_branch.sha} repo={repo} />
          </span>
        </LabeledValue>
      ) : null}
      <ResourceOutcomes run={run} />
      {workflow ? <DeploymentAlerts /> : null}
      <Tabs
        naturalHeight
        initActiveTab={run.status === 'success' ? 'changes' : 'workflow'}
        tabsClassName="mt-4"
        tabControlsClassName="!gap-2 md:!gap-6 [&>button]:px-1 md:[&>button]:px-3"
        tabLabels={Object.fromEntries(
          Object.entries(DEPLOYMENT_TABS).map(([key, tab]) => [key, tab.text])
        )}
        tabs={Object.fromEntries(
          tabOrder.map((key) => [
            key,
            <div key={key} className="flex flex-col gap-4">
              <div className="flex justify-end">
                <Link
                  href={`${basePath}${DEPLOYMENT_TABS[key].path === '/' ? '' : DEPLOYMENT_TABS[key].path}${search}`}
                >
                  Open full page
                </Link>
              </div>
              {content[key]}
            </div>,
          ])
        )}
      />
    </>
  )
}

const WorkflowDeploymentPanelContent = (props: IDeploymentDetailPanel) => {
  const { workflow } = useWorkflow()
  return <DeploymentPanelContent {...props} workflow={workflow} />
}

export const DeploymentDetailPanel = (props: IDeploymentDetailPanel) => {
  const { deployment, orgId, appId, installId, repo, ...panelProps } = props
  const detailProps = { deployment, orgId, appId, installId, repo }
  return (
    <Panel
      heading={deployment.title}
      size="3/4"
      aria-label="Deployment details"
      {...panelProps}
    >
      <WorkflowProvider workflowId={deployment.id} shouldPoll>
        <WorkflowDeploymentPanelContent {...detailProps} />
      </WorkflowProvider>
    </Panel>
  )
}
