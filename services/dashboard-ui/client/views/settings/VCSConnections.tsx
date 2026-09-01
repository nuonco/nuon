import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ConnectGithubButton } from '@/components/vcs-connections/ConnectGithub'
import { VCSConnectionsTable } from '@/components/vcs-connections/VCSConnectionsTable'
import { useOrg } from '@/hooks/use-org'

export const VCSConnections = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="VCS connections" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          { path: `/${org?.id}/settings/vcs`, text: 'VCS connections' },
        ]}
      />
      <ListPage
        title="VCS connections"
        description="Connect GitHub accounts so Nuon can build components from your repositories."
        createAction={
          <ConnectGithubButton variant="primary">
            Add connection
          </ConnectGithubButton>
        }
      >
        <VCSConnectionsTable />
      </ListPage>
    </>
  )
}
