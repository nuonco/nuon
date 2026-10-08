import { DeploymentWorkflowContent } from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import { PageTitle } from '@/components/navigation/PageTitle'
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
      <DeploymentWorkflowContent />
    </>
  )
}
