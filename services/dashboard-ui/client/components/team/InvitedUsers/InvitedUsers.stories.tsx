export default {
  title: 'Team/InvitedUsers',
}

import { InvitedUsers } from './InvitedUsers'
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

const roleTitles = (roleType: string | undefined) =>
  (({ org_admin: 'Admin', org_support: 'Support' }) as Record<string, string>)[
    roleType ?? ''
  ] ??
  roleType ??
  '—'

export const Default = () => (
  <InvitedUsers
    invites={mockInvites}
    roleTitles={roleTitles}
    isLoading={false}
    isError={false}
  />
)

export const Empty = () => (
  <InvitedUsers
    invites={[]}
    roleTitles={roleTitles}
    isLoading={false}
    isError={false}
  />
)

export const WithAcceptedFiltered = () => (
  <InvitedUsers
    invites={[
      ...mockInvites,
      {
        id: 'inv-3',
        email: 'done@example.com',
        status: 'accepted',
        role_type: 'org_admin',
      } as TOrgInvite,
    ]}
    roleTitles={roleTitles}
    isLoading={false}
    isError={false}
  />
)

export const Error = () => (
  <InvitedUsers
    invites={[]}
    roleTitles={roleTitles}
    isLoading={false}
    isError={true}
  />
)

export const Loading = () => (
  <InvitedUsers
    invites={[]}
    roleTitles={roleTitles}
    isLoading
    isError={false}
  />
)
