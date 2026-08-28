import { ReleasesTable } from '@/components/apps/bundles/ReleasesTable'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Releases = () => {
  const { app } = useApp()
  const { org } = useOrg()

  return (
    <>
      <PageTitle segments={['Releases', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/releases`, text: 'Releases' },
        ]}
      />
      <ListPage
        title="Releases"
        description="View immutable app configurations prepared for customer-managed installs."
      >
        <ReleasesTable />
      </ListPage>
    </>
  )
}
