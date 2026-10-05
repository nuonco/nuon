import { PageTitle } from '@/components/navigation/PageTitle'
import { WorkflowChangesSummaryContainer } from '@/components/workflows/WorkflowChangesSummary'
import { useInstallPage } from '@/hooks/use-install-path'
import { useWorkflow } from '@/hooks/use-workflow'
import { humanize } from '@/utils/string-utils'

export const DeploymentChangesTab = () => {
  const { install } = useInstallPage()
  const { workflow } = useWorkflow()
  const title = workflow?.name || humanize(workflow?.type) || 'Deployment'

  return (
    <>
      <PageTitle segments={[title, install?.name]} />
      <WorkflowChangesSummaryContainer />
    </>
  )
}
