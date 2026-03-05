import { useParams } from 'react-router'
import { BackToTop } from '@/components/common/BackToTop'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { WorkflowDetails } from '@/components/workflows/WorkflowDetails'
import { WorkflowSteps, WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { useWorkflow } from '@/hooks/use-workflow'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { snakeToWords, toSentenceCase } from '@/utils/string-utils'

const CONTAINER_ID = 'workflow-page'

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

  const workflowName =
    workflow?.name || toSentenceCase(snakeToWords(workflow?.type)) || 'Workflow'

  return (
    <PageSection id={CONTAINER_ID} isScrollable className="!gap-2 !pb-24">
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: `/${org?.id}/installs/${install?.id}/workflows`, text: 'Workflows' },
          {
            path: `/${org?.id}/installs/${install?.id}/workflows/${workflowId}`,
            text: workflowName,
          },
        ]}
      />

      <WorkflowDetails />

      <div className="flex flex-col gap-6 mt-6">
        <Text variant="h3" weight="strong">
          Workflow steps
        </Text>

        {workflow ? (
          <WorkflowSteps
            approvalPrompt={workflow?.approval_option === 'prompt'}
            planOnly={workflow?.plan_only}
            workflowId={workflowId!}
            shouldPoll
          />
        ) : (
          <WorkflowStepsSkeleton />
        )}
      </div>

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
