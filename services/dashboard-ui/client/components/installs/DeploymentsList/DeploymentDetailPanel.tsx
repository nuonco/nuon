import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { DeploymentDetailLoading } from './DeploymentDetailLoading'
import {
  DeploymentAlerts,
  DeploymentChangesContent,
  DeploymentTemplateContent,
  DeploymentWorkflowContent,
} from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import {
  useDeploymentRunBranch,
  DeploymentMeta,
} from '@/components/installs/DeploymentDetail/DeploymentMeta'
import {
  DeploymentRunStatus,
  ResourceScopeSummary,
} from '@/components/installs/DeploymentDetail/DeploymentProgress'
import {
  DEPLOYMENT_TABS,
  type TDeploymentTab,
  deploymentOutcomes,
  deploymentSteps,
  deploymentTabOrder,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { useWorkflow } from '@/hooks/use-workflow'
import { getInstallDeployment } from '@/lib'
import { WorkflowProvider } from '@/providers/workflow-provider'
import type { TInstallDeploymentSummary, TWorkflow } from '@/types'

export interface IDeploymentDetailPanel extends IPanel {
  deployment: TInstallDeploymentSummary
  orgId: string
  appId: string
  installId: string
  repo?: string
  initialTab?: TDeploymentTab
}

const DeploymentPanelContent = ({
  deployment,
  orgId,
  appId,
  installId,
  repo,
  initialTab,
  workflow,
}: IDeploymentDetailPanel & { workflow?: TWorkflow }) => {
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
  const type = record?.type || deployment.type || workflow?.type
  const branch = useDeploymentRunBranch({
    orgId,
    appId,
    type,
    installId,
    deployment: record,
    workflow,
    repo,
  })
  const run = {
    status: workflow?.status?.status ?? deployment.status,
    activity:
      workflow?.status?.status_human_description ?? deployment.activity ?? '',
    steps: deploymentSteps(workflow),
  }
  const [tabOrder] = useState(() => deploymentTabOrder(workflow?.type))
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
        installId={installId}
        workflow={workflow}
        isLoading={recordPending}
        isError={recordError}
        onRetry={() => refetchRecord()}
      />
    ),
    workflow: workflow ? (
      <DeploymentWorkflowContent installId={installId} />
    ) : (
      noWorkflow
    ),
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
      {run.activity && run.status !== 'error' ? (
        <Text variant="subtext" theme="neutral">
          {run.activity}
        </Text>
      ) : null}
      <DeploymentMeta framed type={type} branch={branch} />
      <ResourceScopeSummary
        run={{
          status: run.status,
          outcomes: deploymentOutcomes(record, workflow),
        }}
      />
      {workflow ? <DeploymentAlerts /> : null}
      <Tabs
        naturalHeight
        initActiveTab={initialTab ?? tabOrder[0]}
        tabsClassName="mt-4"
        tabControlsClassName="!gap-2 md:!gap-6 [&>button]:px-1 md:[&>button]:px-3"
        tabLabels={Object.fromEntries(
          Object.entries(DEPLOYMENT_TABS).map(([key, tab]) => [key, tab.text])
        )}
        tabs={Object.fromEntries(
          tabOrder.map((key) => [
            key,
            <div key={key} className="flex flex-col gap-4">
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

export type TDeploymentDetailBody = Pick<
  IDeploymentDetailPanel,
  'deployment' | 'orgId' | 'appId' | 'installId' | 'repo' | 'initialTab'
>

export const DeploymentDetailBody = (props: TDeploymentDetailBody) => (
  <WorkflowProvider
    workflowId={props.deployment.id}
    shouldPoll
    loadingElement={
      <DeploymentDetailLoading
        activity={props.deployment.activity}
        createdAt={props.deployment.created_at}
        id={props.deployment.id}
        status={props.deployment.status}
        workflowType={props.deployment.type}
        initialTab={props.initialTab}
      />
    }
  >
    <WorkflowDeploymentPanelContent {...props} />
  </WorkflowProvider>
)

export const DeploymentDetailPanel = (props: IDeploymentDetailPanel) => {
  const {
    deployment,
    orgId,
    appId,
    installId,
    repo,
    initialTab,
    ...panelProps
  } = props
  return (
    <Panel
      heading={deployment.title}
      size="3/4"
      aria-label="Deployment details"
      {...panelProps}
    >
      <DeploymentDetailBody
        deployment={deployment}
        orgId={orgId}
        appId={appId}
        installId={installId}
        repo={repo}
        initialTab={initialTab}
      />
    </Panel>
  )
}
