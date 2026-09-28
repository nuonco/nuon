import { ConnectionsList } from '@/components/cloud-connections/ConnectionsList'
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
      <ConnectionsList />
    </>
  )
}
