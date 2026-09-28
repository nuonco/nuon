import { useParams } from 'react-router'
import { DetailPage } from '@/components/layout/DetailPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { WorkflowDetails } from '@/components/workflows/WorkflowDetails'
import {
  WorkflowSteps,
  WorkflowStepsSkeleton,
} from '@/components/workflows/WorkflowSteps'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { useWorkflow } from '@/hooks/use-workflow'
import { useInstallNested } from '@/hooks/use-install-path'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { humanize } from '@/utils/string-utils'

export const WorkflowDetail = () => {
  const { workflowId } = useParams()

  return (
    <WorkflowProvider workflowId={workflowId!} shouldPoll>
      <WorkflowDetailContent />
    </WorkflowProvider>
  )
}

const WorkflowDetailContent = () => {
  const { workflowId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const { workflow } = useWorkflow()
  const nested = useInstallNested()

  const workflowName = workflow?.name || humanize(workflow?.type) || 'Workflow'
  const installRoot = `/${org?.id}/installs/${install?.id}`
  const section = nested
    ? { path: `${installRoot}/deployments`, text: 'Deployments' }
    : { path: `${installRoot}/workflows`, text: 'Workflows' }
  const page = nested
    ? `${installRoot}/deployments/${workflowId}`
    : `${installRoot}/workflows/${workflowId}`

  return (
    <>
      <PageTitle segments={[workflowName, install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: installRoot, text: install?.name },
          section,
          { path: page, text: workflowName },
        ]}
      />

      <DetailPage header={<WorkflowDetails />}>
        <div className="flex flex-col gap-6">
          <SectionHeader title="Workflow steps" />

          {workflow ? (
            <WorkflowSteps
              approvalPrompt={workflow?.approval_option === 'prompt'}
              planOnly={workflow?.plan_only}
            />
          ) : (
            <WorkflowStepsSkeleton />
          )}
        </div>
      </DetailPage>
    </>
  )
}
