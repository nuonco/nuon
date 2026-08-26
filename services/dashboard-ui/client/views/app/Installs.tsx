import { InstallsTable } from '@/components/installs/InstallsTable'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Installs = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <>
      <PageTitle segments={['Installs', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/installs`, text: 'Installs' },
        ]}
      />
      <ListPage
        title="App installs"
        description="View and manage deployments of your app into customer cloud accounts."
      >
        <InstallsTable appId={app?.id} />
      </ListPage>
    </>
  )
}
