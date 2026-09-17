import { DateTime } from 'luxon'
import type { TOrgMember, TRoleInfo } from '@/types/ctl-api.types'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { TeamTable, type ITeamFilter } from './TeamTable'

export default {
  title: 'lite/organisms/TeamTable',
}

const ROLES: TRoleInfo[] = [
  {
    id: 'role_admin',
    role_type: 'org_admin',
    title: 'Org admin',
    description: 'Full access to org settings and resources.',
  },
  {
    id: 'role_read_only',
    role_type: 'org_read_only',
    title: 'Read only',
    description: 'View access to org resources.',
  },
]

const MEMBERS: TOrgMember[] = [
  {
    id: 'acct_current',
    account_id: 'acct_current',
    email: 'current.user@example.com',
    name: 'Current User',
    status: 'active',
    role_type: 'org_admin',
    joined_at: DateTime.now().minus({ months: 8 }).toISO(),
    created_at: DateTime.now().minus({ months: 8 }).toISO(),
  },
  {
    id: 'acct_engineer',
    account_id: 'acct_engineer',
    email: 'engineer@example.com',
    name: 'Example Engineer',
    status: 'active',
    role_type: 'org_read_only',
    joined_at: DateTime.now().minus({ days: 12 }).toISO(),
    created_at: DateTime.now().minus({ days: 12 }).toISO(),
  },
  {
    id: 'invite_operator',
    invite_id: 'invite_operator',
    email: 'operator@example.com',
    status: 'invited',
    role_type: 'org_read_only',
    created_at: DateTime.now().minus({ days: 2 }).toISO(),
  },
]

const filter = (
  label: string,
  options: { value: string; label: string }[]
): ITeamFilter => ({
  label,
  options,
  selected: new Set(),
  onToggle: () => {},
  onIsolate: () => {},
  onReset: () => {},
  constrained: false,
})

const props = {
  members: MEMBERS,
  roles: ROLES,
  currentAccountId: 'acct_current',
  admin: true,
  search: '',
  onSearchChange: () => {},
  offset: 0,
  pageSize: 20,
  hasNext: false,
  onOffsetChange: () => {},
  statusFilter: filter('Status', [
    { value: 'active', label: 'Active' },
    { value: 'invited', label: 'Invited' },
  ]),
  roleFilter: filter(
    'Role',
    ROLES.map((role) => ({
      value: role.role_type,
      label: role.title || role.role_type,
    }))
  ),
  onChangeRole: () => {},
  onRemove: () => {},
  onResend: () => {},
  onRevoke: () => {},
}

export const Overview = () => (
  <ComponentDocs
    name="TeamTable"
    tier="organism"
    summary="The org's active members and pending invites in one searchable, filterable, paginated collection."
    use={[
      'Render it through TeamTableContainer on the Team page.',
      'Use the card presentation when the available width cannot fit the table columns.',
    ]}
    avoid={[
      'Do not split active members and pending invites into separate collections.',
      'Do not filter, sort, merge, or paginate resolved members in the browser.',
      'Do not infer whether a row is active or invited from status when account_id and invite_id provide that distinction.',
    ]}
    rules={[
      'Email is the primary identity; name supports recognition and avatar initials.',
      'Role titles come from the team role query and fall back to role_type.',
      'Status is passed directly to Status without mapping a theme at the call site.',
      'Change role and revoke invite are hidden from non-admins, and change role is hidden on the current account row.',
      'Search and filter changes reset the URL-backed offset in the container.',
      'Pagination sits below Table and remains available in both table and card presentations.',
    ]}
    props={[
      {
        name: 'members',
        type: 'TOrgMember[]',
        description:
          'Resolved members and pending invites for the current page.',
      },
      {
        name: 'roles',
        type: 'TRoleInfo[]',
        description: 'Team roles used to resolve role titles.',
      },
      {
        name: 'currentAccountId',
        type: 'string',
        description: 'Current account used to hide its change-role action.',
      },
      {
        name: 'admin',
        type: 'boolean',
        description: 'Whether org-admin-only row actions are shown.',
      },
      {
        name: 'search',
        type: 'string',
        description: 'Current search term from the URL.',
      },
      {
        name: 'onSearchChange',
        type: '(value: string) => void',
        description: 'Writes the search term and resets the offset.',
      },
      {
        name: 'offset',
        type: 'number',
        description: 'Row offset of the current page.',
      },
      {
        name: 'pageSize',
        type: 'number',
        description: 'Rows requested per page.',
      },
      {
        name: 'hasNext',
        type: 'boolean',
        description: 'Whether another page is available.',
      },
      {
        name: 'onOffsetChange',
        type: '(offset: number) => void',
        description: 'Moves the page window.',
      },
      {
        name: 'statusFilter',
        type: 'ITeamFilter',
        description:
          'Active and invited status options with URL-backed selection.',
      },
      {
        name: 'roleFilter',
        type: 'ITeamFilter',
        description: 'Available role options with URL-backed selection.',
      },
      {
        name: 'onChangeRole',
        type: '(member: TOrgMember) => void',
        description: 'Starts the change-role action for an active member.',
      },
      {
        name: 'onRemove',
        type: '(member: TOrgMember) => void',
        description: 'Starts the remove action for an active member.',
      },
      {
        name: 'onResend',
        type: '(member: TOrgMember) => void',
        description: 'Starts the resend action for a pending invite.',
      },
      {
        name: 'onRevoke',
        type: '(member: TOrgMember) => void',
        description: 'Starts the revoke action for a pending invite.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders Table skeletons during the cold load.',
      },
      {
        name: 'fetching',
        type: 'boolean',
        default: 'false',
        description: 'Disables pagination while a page is in flight.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Swaps the empty state to failure copy.',
      },
    ]}
  />
)

export const Default = () => <TeamTable {...props} />

export const AllActive = () => (
  <TeamTable
    {...props}
    members={MEMBERS.filter((member) => member.account_id)}
  />
)

export const AllInvited = () => (
  <TeamTable
    {...props}
    members={MEMBERS.filter((member) => member.invite_id)}
  />
)

export const Loading = () => <TeamTable {...props} members={[]} loading />

export const Empty = () => <TeamTable {...props} members={[]} />

export const Error = () => (
  <TeamTable {...props} members={[]} error={new globalThis.Error()} />
)

export const Cards = () => (
  <div className="max-w-lg p-4">
    <TeamTable {...props} />
  </div>
)

export const SelfRow = () => <TeamTable {...props} members={[MEMBERS[0]]} />

export const NonAdmin = () => <TeamTable {...props} admin={false} />

export const LongEmails = () => (
  <TeamTable
    {...props}
    members={[
      {
        ...MEMBERS[1],
        email:
          'platform-engineering-and-infrastructure-operations@example.example.com',
      },
    ]}
  />
)
