import { InstallsTable } from '@/components/installs/InstallsTable'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const Installs = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="Installs" />
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
          {
            path: `/${org.id}/installs`,
            text: 'Installs',
          },
        ]}
      />
      <ListPage
        variant="page"
        title="Installs"
        description="View and manage all deployed installs here."
      >
        <InstallsTable shouldPoll />
      </ListPage>
    </>
  )
}
