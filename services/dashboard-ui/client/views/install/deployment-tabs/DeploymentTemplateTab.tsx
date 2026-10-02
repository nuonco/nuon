import { useOutletContext } from 'react-router'
import { BranchRunChanges } from '@/components/branches/BranchRunChanges/BranchRunChanges'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstallPage } from '@/hooks/use-install-path'
import { useWorkflow } from '@/hooks/use-workflow'
import { AppProvider } from '@/providers/app-provider'
import type { TInstallDeploymentRecord } from '@/types'
import { humanize } from '@/utils/string-utils'

export const DeploymentTemplateTab = () => {
  const { install } = useInstallPage()
  const { workflow } = useWorkflow()
  const { deployment, deploymentFetched } = useOutletContext<{
    deployment?: TInstallDeploymentRecord
    deploymentFetched: boolean
  }>()
  const title =
    deployment?.title ||
    workflow?.name ||
    humanize(workflow?.type) ||
    'Deployment'
  const branch = deployment?.app_branch

  return (
    <>
      <PageTitle segments={[title, install?.name]} />
      {!deploymentFetched ? (
        <Text loading loadingWidth={32} />
      ) : branch?.id && branch.run_id && install?.app_id ? (
        <AppProvider appId={install.app_id}>
          <BranchRunChanges
            branchId={branch.id}
            appBranchRunId={branch.run_id}
            showRunComparison={false}
            title="Template updates"
          />
        </AppProvider>
      ) : (
        <EmptyState
          emptyTitle="No template updates"
          emptyMessage="This deployment has no app config changes."
        />
      )}
    </>
  )
}
