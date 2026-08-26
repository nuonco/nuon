import { ActionsTable } from '@/components/actions/ActionsTable'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Actions = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <>
      <PageTitle segments={['Actions', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/actions`, text: 'Actions' },
        ]}
      />
      <ListPage
        title="App actions"
        description="Configure and run day-2 operations on your installs."
      >
        <ActionsTable />
      </ListPage>
    </>
  )
}
