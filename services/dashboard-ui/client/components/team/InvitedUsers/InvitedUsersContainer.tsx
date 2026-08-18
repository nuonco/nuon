import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useRoleTitles } from '@/hooks/use-roles'
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
  const roleTitles = useRoleTitles()

  const { data: invites, isLoading, isError } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['org-invites', org?.id],
    queryFn: () => getOrgInvites({ orgId: org.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id,
  })

  return (
    <InvitedUsers
      invites={invites ?? []}
      roleTitles={roleTitles}
      isLoading={isLoading}
      isError={isError}
    />
  )
}
