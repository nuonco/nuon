import {
  CloudConnectionsTable,
  CreateCloudConnectionButton,
} from '@/components/cloud-connections'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const CloudConnections = () => {
  const { org } = useOrg()
  return (
    <>
      <PageTitle title="Cloud connections" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/cloud-connections`, text: 'Cloud connections' },
        ]}
      />
      <ListPage
        variant="page"
        title="Cloud connections"
        description="Connect cloud accounts so Nuon can deploy install stacks and pull private images."
        actions={<CreateCloudConnectionButton />}
      >
        <CloudConnectionsTable />
      </ListPage>
    </>
  )
}
