import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getOrgInvites } from '@/lib'
import { InvitedUsers } from './InvitedUsers'

export const InvitedUsersContainer = ({
  shouldPoll = false,
  pollInterval = 20000,
}: {
  shouldPoll?: boolean
  pollInterval?: number
}) => {
  const { org } = useOrg()

  const { data: invites, isLoading, isError } = useQuery({
    queryKey: ['org-invites', org?.id],
    queryFn: () => getOrgInvites({ orgId: org.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id,
  })

  return (
    <InvitedUsers
      invites={invites ?? []}
      isLoading={isLoading}
      isError={isError}
    />
  )
}
