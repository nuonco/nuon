import { ComponentsTable } from '@/components/components/ComponentsTable'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Components = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <>
      <PageTitle segments={['Components', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/components`, text: 'Components' },
        ]}
      />
      <ListPage
        title="App components"
        description="Manage the components that make up your application."
      >
        <div className="flex flex-auto min-w-0">
          <ComponentsTable />
        </div>
      </ListPage>
    </>
  )
}
