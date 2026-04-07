export default {
  title: 'Team/InvitedUsers',
}

import { InvitedUsers, InvitedUsersSkeleton } from './InvitedUsers'
import type { TOrgInvite } from '@/types'

const mockInvites: TOrgInvite[] = [
  {
    id: 'inv-1',
    email: 'pending@example.com',
    status: 'pending',
    role_type: 'org_admin',
  } as TOrgInvite,
  {
    id: 'inv-2',
    email: 'waiting@example.com',
    status: 'pending',
    role_type: 'org_support',
  } as TOrgInvite,
]

export const Default = () => (
  <InvitedUsers invites={mockInvites} isLoading={false} isError={false} />
)

export const Empty = () => (
  <InvitedUsers invites={[]} isLoading={false} isError={false} />
)

export const WithAcceptedFiltered = () => (
  <InvitedUsers
    invites={[
      ...mockInvites,
      { id: 'inv-3', email: 'done@example.com', status: 'accepted', role_type: 'org_admin' } as TOrgInvite,
    ]}
    isLoading={false}
    isError={false}
  />
)

export const Error = () => (
  <InvitedUsers invites={[]} isLoading={false} isError={true} />
)

export const Loading = () => <InvitedUsersSkeleton />
