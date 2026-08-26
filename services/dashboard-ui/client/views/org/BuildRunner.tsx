import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ControlPlaneRecentActivity } from '@/components/runners/ControlPlaneRecentActivity'
import { useOrg } from '@/hooks/use-org'

export const BuildRunner = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="Builds" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/runner`, text: 'Builds' },
        ]}
      />
      <ListPage
        variant="page"
        title="Builds"
        description="Component builds run on Nuon's control plane."
      >
        <ControlPlaneRecentActivity shouldPoll />
      </ListPage>
    </>
  )
}
