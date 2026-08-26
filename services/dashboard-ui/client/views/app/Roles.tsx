import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { AppRolesTable } from '@/components/roles/AppRolesTable'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Roles = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <>
      <PageTitle segments={['Roles', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/roles`, text: 'Roles' },
        ]}
      />
      <ListPage
        title="App roles"
        description="View the IAM roles that your app uses to access customer cloud resources."
      >
        <AppRolesTable />
      </ListPage>
    </>
  )
}
