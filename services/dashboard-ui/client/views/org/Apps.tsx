import { AppsTable } from '@/components/apps/AppsTable'
import { CreateAppButton } from '@/components/apps/CreateAppModal'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const Apps = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="Apps" />
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
          {
            path: `/${org.id}/apps`,
            text: 'Apps',
          },
        ]}
      />
      <ListPage
        variant="page"
        title="Apps"
        description="Manage your applications here."
        createAction={
          <CreateAppButton variant="primary" />
        }
      >
        <AppsTable shouldPoll />
      </ListPage>
    </>
  )
}
