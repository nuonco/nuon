import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageTitle } from '@/components/navigation/PageTitle'
import { WorkflowDetails } from '@/components/workflows/WorkflowDetails'
import {
  WorkflowSteps,
  WorkflowStepsSkeleton,
} from '@/components/workflows/WorkflowSteps'
import { useInstallPage } from '@/hooks/use-install-path'
import { useWorkflow } from '@/hooks/use-workflow'
import { humanize } from '@/utils/string-utils'

export const DeploymentWorkflowTab = () => {
  const { install } = useInstallPage()
  const { workflow } = useWorkflow()
  const title = workflow?.name || humanize(workflow?.type) || 'Deployment'

  return (
    <>
      <PageTitle segments={[title, install?.name]} />
      <div className="flex flex-col gap-6">
        <WorkflowDetails showBanners={false} />
        <div className="flex flex-col gap-4">
          <SectionHeader title="Workflow steps" />
          {workflow ? (
            <WorkflowSteps
              approvalPrompt={workflow.approval_option === 'prompt'}
              planOnly={workflow.plan_only}
            />
          ) : (
            <WorkflowStepsSkeleton />
          )}
        </div>
      </div>
    </>
  )
}
