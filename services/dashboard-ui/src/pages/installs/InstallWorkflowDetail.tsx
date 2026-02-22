import { useParams, useSearchParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { BackToTop } from '@/components/common/BackToTop'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { WorkflowDetails } from '@/components/workflows/WorkflowDetails'
import { WorkflowSteps, WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { snakeToWords, toSentenceCase } from '@/utils/string-utils'
import type { TWorkflow, TWorkflowStep } from '@/types'

const WorkflowStepsWrapper = ({
  workflowId,
  approvalPrompt,
  planOnly,
}: {
  workflowId: string
  approvalPrompt: boolean
  planOnly: boolean
}) => {
  const [searchParams] = useSearchParams()
  const offset = searchParams.get('offset') || '0'

  const {
    data: steps,
    error,
    isLoading,
  } = usePolling<TWorkflowStep[]>({
    path: `/api/ctl-api/v1/workflows/${workflowId}/steps?offset=${offset}`,
    shouldPoll: true,
    pollInterval: 4000,
  })

  if (error) {
    return <Text>Error fetching workflow steps</Text>
  }

  if (isLoading && !steps) {
    return <WorkflowStepsSkeleton />
  }

  return (
    <WorkflowSteps
      approvalPrompt={approvalPrompt}
      initWorkflowSteps={steps || []}
      planOnly={planOnly}
      shouldPoll
      workflowId={workflowId}
    />
  )
}

export default function InstallWorkflowDetail() {
  const { orgId, installId, workflowId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: workflow,
    error,
    isLoading,
  } = usePolling<TWorkflow>({
    path: `/api/ctl-api/v1/workflows/${workflowId}`,
    shouldPoll: true,
    pollInterval: 5000,
  })

  if (!workflowId || !installId || !orgId) {
    return null
  }

  if (error) {
    return <Text>Workflow not found</Text>
  }

  if (isLoading && !workflow) {
    return <div>Loading workflow...</div>
  }

  if (!workflow) {
    return <div>Loading workflow...</div>
  }

  const workflowName =
    workflow?.name || snakeToWords(toSentenceCase(workflow?.type))
  const containerId = 'workflow-page'

  return (
    <PageSection id={containerId} isScrollable className="!gap-2 !pb-24">
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/workflows`,
            text: 'Workflows',
          },
          {
            path: `/${orgId}/installs/${installId}/workflows/${workflowId}`,
            text: workflowName,
          },
        ]}
      />
      <WorkflowProvider initWorkflow={workflow} shouldPoll>
        <WorkflowDetails />

        <div className="flex flex-col gap-6 mt-6">
          <Text variant="h3" weight="strong">
            Workflow steps
          </Text>
          <WorkflowStepsWrapper
            workflowId={workflowId}
            approvalPrompt={workflow?.approval_option === 'prompt'}
            planOnly={workflow?.plan_only || false}
          />
        </div>
      </WorkflowProvider>
      <BackToTop containerId={containerId} />
    </PageSection>
  )
}
