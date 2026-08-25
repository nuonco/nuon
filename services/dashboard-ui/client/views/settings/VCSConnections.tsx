import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
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
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            VCS connections
          </Text>
          <Text theme="neutral">
            Connect GitHub accounts so Nuon can build components from your
            repositories.
          </Text>
        </HeadingGroup>
        <ConnectGithubButton isIconFirst variant="primary">
          Add connection
        </ConnectGithubButton>
      </PageHeader>
      <PageContent>
        <PageSection>
          <VCSConnectionsTable />
        </PageSection>
      </PageContent>
    </>
  )
}
