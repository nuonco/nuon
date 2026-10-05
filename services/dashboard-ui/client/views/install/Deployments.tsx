import { DeploymentsList } from '@/components/installs/DeploymentsList'
import { InstallWorkflowPanelController } from '@/components/workflows/InstallWorkflowPanel'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Deployments = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Deployments', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/deployments`,
            text: 'Deployments',
          },
        ]}
      />
      <ListPage
        title="Deployments"
        description={`Follow rollouts for ${install?.name}. View details to inspect a workflow or its changes.`}
      >
        <DeploymentsList shouldPoll />
      </ListPage>
      <InstallWorkflowPanelController />
    </>
  )
}
