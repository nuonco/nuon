import { useParams } from 'react-router'
import { CompositeError } from '@/components/common/CompositeError'
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

  const workflowName = workflow?.name || humanize(workflow?.type) || 'Workflow'

  return (
    <>
      <PageTitle segments={[workflowName, install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/workflows`,
            text: 'Workflows',
          },
          {
            path: `/${org?.id}/installs/${install?.id}/workflows/${workflowId}`,
            text: workflowName,
          },
        ]}
      />

      <DetailPage
        header={<WorkflowDetails />}
        banners={
          workflow?.preflight_errors?.length ? (
            <div className="flex flex-col gap-2">
              {workflow.preflight_errors.map((error, index) => (
                <CompositeError
                  key={`${error?.type ?? 'preflight'}-${index}`}
                  error={error}
                />
              ))}
            </div>
          ) : null
        }
      >
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
