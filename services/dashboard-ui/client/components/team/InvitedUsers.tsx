import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ResendOrgInviteButton } from '@/components/team/ResendOrgInvite'
import { useOrg } from '@/hooks/use-org'
import { getOrgInvites } from '@/lib'

export const InvitedUsers = ({
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

  if (isLoading) return <InvitedUsersSkeleton />
  if (isError) return <InvitedUsersError />

  const pendingInvites = invites?.filter((i) => i?.status !== 'accepted') ?? []

  if (!pendingInvites.length) {
    return (
      <InvitedUsersError
        title="No active invites"
        message="No outstanding invites to this org"
      />
    )
  }

  return (
    <div className="flex flex-col gap-2">
      {pendingInvites.map((i) => (
        <div className="flex items-center gap-4" key={i?.id}>
          <Status variant="badge" status={i?.status} />
          <Text variant="subtext">{i?.email}</Text>
          <Badge size="sm" variant="code">
            {i?.role_type === 'org_admin' ? 'Admin' : i?.role_type}
          </Badge>
          <ResendOrgInviteButton invite={i} size="sm" />
        </div>
      ))}
    </div>
  )
}

export const InvitedUsersError = ({
  message = 'We encountered an issue loading invites. Please try refreshing the page.',
  title = 'Unable to load user invites',
}: {
  message?: string
  title?: string
}) => {
  return <EmptyState variant="table" emptyMessage={message} emptyTitle={title} />
}

export const InvitedUsersSkeleton = () => {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-4">
        <Skeleton height="23px" width="75px" />
        <Skeleton height="17px" width="110px" />
        <Skeleton height="20px" width="50px" />
      </div>
      <div className="flex items-center gap-4">
        <Skeleton height="23px" width="75px" />
        <Skeleton height="17px" width="110px" />
        <Skeleton height="20px" width="50px" />
      </div>
    </div>
  )
}
