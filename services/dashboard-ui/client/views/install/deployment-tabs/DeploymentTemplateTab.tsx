import { useOutletContext } from 'react-router'
import { DeploymentTemplateContent } from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstallPage } from '@/hooks/use-install-path'
import { useWorkflow } from '@/hooks/use-workflow'
import type { TInstallDeploymentRecord } from '@/types'
import { humanize } from '@/utils/string-utils'

export const DeploymentTemplateTab = () => {
  const { install } = useInstallPage()
  const { workflow } = useWorkflow()
  const { deployment, deploymentFetched, deploymentError, refetchDeployment } =
    useOutletContext<{
      deployment?: TInstallDeploymentRecord
      deploymentFetched: boolean
      deploymentError: boolean
      refetchDeployment: () => void
    }>()
  const title =
    deployment?.title ||
    workflow?.name ||
    humanize(workflow?.type) ||
    'Deployment'

  return (
    <>
      <PageTitle segments={[title, install?.name]} />
      <DeploymentTemplateContent
        deployment={deployment}
        appId={install?.app_id}
        workflow={workflow}
        isLoading={!deploymentFetched}
        isError={deploymentError}
        onRetry={refetchDeployment}
      />
    </>
  )
}
