import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TeamTable } from '@/components/team/TeamTable'
import { InviteUserButton } from '@/components/team/InviteUser'
import { InvitedUsers } from '@/components/team/InvitedUsers'

import { useOrg } from '@/hooks/use-org'

export const Team = () => {
  const { org } = useOrg()

  return (
    <PageLayout className="pb-6">
      <PageTitle title={`Team | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
          {
            path: `/${org.id}/team`,
            text: 'Team',
          },
        ]}
      />
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Team
          </Text>
          <Text theme="neutral">Manage your team members and permissions.</Text>
        </HeadingGroup>
        <InviteUserButton />
      </PageHeader>
      <PageContent>
        <PageSection>
          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Active members
            </Text>
            <TeamTable shouldPoll />
          </div>

          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Active invites
            </Text>
            <InvitedUsers shouldPoll />
          </div>
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}
