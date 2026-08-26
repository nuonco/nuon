import { Text } from '@/components/common/Text'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TeamTable } from '@/components/team/TeamTable'
import { InviteUserButton } from '@/components/team/InviteUser'
import { InvitedUsers } from '@/components/team/InvitedUsers'

import { useOrg } from '@/hooks/use-org'

export const Team = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="Team" />
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
      <ListPage
        variant="page"
        title="Team"
        description="Manage your team members and permissions."
        createAction={<InviteUserButton variant="primary" />}
      >
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
      </ListPage>
    </>
  )
}
