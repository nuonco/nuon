import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ResendOrgInviteButton } from '@/components/team/ResendOrgInvite'
import { RevokeOrgInviteButton } from '@/components/team/RevokeOrgInvite'
import type { TOrgInvite } from '@/types'

export const InvitedUsers = ({
  invites,
  roleTitles,
  isLoading,
  isError,
}: {
  invites: TOrgInvite[]
  roleTitles: (roleType: string | undefined) => string
  isLoading: boolean
  isError: boolean
}) => {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-2">
        {Array.from({ length: 2 }).map((_, i) => (
          <div key={i} className="flex items-center gap-4">
            <Status loading variant="badge" />
            <Text variant="subtext" loading loadingWidth={18} />
            <Badge loading size="sm" variant="code" />
          </div>
        ))}
      </div>
    )
  }
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
            {roleTitles(i?.role_type)}
          </Badge>
          <ResendOrgInviteButton invite={i} size="sm" />
          <RevokeOrgInviteButton invite={i} size="sm" />
        </div>
      ))}
    </div>
  )
}

export const InvitedUsersError = ({
  message = 'Unable to load invites. Try refreshing the page.',
  title = 'Unable to load user invites',
}: {
  message?: string
  title?: string
}) => {
  return (
    <EmptyState variant="table" emptyMessage={message} emptyTitle={title} />
  )
}
