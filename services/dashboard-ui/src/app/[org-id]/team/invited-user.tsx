import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Status } from '@/components/common/Status'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { ResendOrgInviteButton } from '@/components/team/ResendOrgInvite'
import { getOrgInvites } from '@/lib'

export const InvitedUser = async ({ orgId }: { orgId: string }) => {
  const { data: invites, error, headers } = await getOrgInvites({ orgId })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return error ? (
    <InvitedUserError />
  ) : invites?.length ? (
    <div className="flex flex-col gap-2">
      {invites
        ?.filter((i) => i?.status !== 'accepted')
        ?.map((i) => {
          return (
            <div className="flex items-center gap-4" key={i?.id}>
              <Status variant="badge" status={i?.status} />
              <Text variant="subtext">{i?.email}</Text>
              <Badge size="sm" variant="code">
                {i?.role_type === 'org_admin' ? 'Admin' : i?.role_type}
              </Badge>
              <ResendOrgInviteButton
                invite={i}
                size="sm"
              />
            </div>
          )
        })}
    </div>
  ) : (
    <InvitedUserError
      title="No active invies"
      message="No outstanding invites to this org"
    />
  )
}

export const InvitedUserError = ({
  message = 'We encountered an issue loading invites. Please try refreshing the page.',
  title = 'Unable to load user invites',
}: {
  message?: string
  title?: string
}) => {
  return <EmptyState variant="table" emptyMessage={message} title={title} />
}

export const InvitedUserSkeleton = () => {
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
