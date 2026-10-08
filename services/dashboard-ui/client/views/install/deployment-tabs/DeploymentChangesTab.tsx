import { useOutletContext } from 'react-router'
import { PageTitle } from '@/components/navigation/PageTitle'
import { DeploymentChangesContent } from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import { useInstallPage } from '@/hooks/use-install-path'
import { useWorkflow } from '@/hooks/use-workflow'
import type { TInstallDeploymentRecord } from '@/types'
import { humanize } from '@/utils/string-utils'

export const DeploymentChangesTab = () => {
  const { install } = useInstallPage()
  const { workflow } = useWorkflow()
  const { deployment } = useOutletContext<{
    deployment?: TInstallDeploymentRecord
  }>()
  const title = workflow?.name || humanize(workflow?.type) || 'Deployment'

  return (
    <>
      <PageTitle segments={[title, install?.name]} />
      <DeploymentChangesContent deployment={deployment} />
    </>
  )
}
