import type { ColumnDef } from '@tanstack/react-table'
import type { TOrgMember, TRoleInfo } from '@/types/ctl-api.types'
import { Avatar } from '../../atoms/Avatar'
import { Button } from '../../atoms/Button'
import { Card } from '../../atoms/Card'
import { Dropdown } from '../../atoms/Dropdown'
import { Icon } from '../../atoms/Icon'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import {
  FilterDropdown,
  type IFilterMenuOption,
} from '../../molecules/FilterMenu'
import { ID } from '../../molecules/ID'
import { ListSearch } from '../../molecules/ListSearch'
import { Menu, MenuItem } from '../../molecules/Menu'
import { Pagination } from '../../molecules/Pagination'
import { Time } from '../../molecules/Time'
import { Table } from '../Table/Table'

export interface ITeamFilter {
  label: string
  options: readonly IFilterMenuOption<string>[]
  selected: Set<string>
  onToggle: (value: string) => void
  onIsolate: (value: string) => void
  onReset: () => void
  constrained: boolean
}

export interface ITeamTable {
  members: TOrgMember[]
  roles: TRoleInfo[]
  currentAccountId?: string
  admin: boolean
  search: string
  onSearchChange: (value: string) => void
  offset: number
  pageSize: number
  hasNext: boolean
  onOffsetChange: (offset: number) => void
  statusFilter: ITeamFilter
  roleFilter?: ITeamFilter
  onChangeRole: (member: TOrgMember) => void
  onRemove: (member: TOrgMember) => void
  onResend: (member: TOrgMember) => void
  onRevoke: (member: TOrgMember) => void
  loading?: boolean
  fetching?: boolean
  error?: unknown
}

const roleTitle = (member: TOrgMember, roles: TRoleInfo[]) => {
  const roleType = member?.role_type
  if (!roleType) return '—'
  return roles.find((role) => role.role_type === roleType)?.title ?? roleType
}

const MemberIdentity = ({ member }: { member: TOrgMember }) => (
  <span className="flex min-w-0 items-center gap-3">
    <Avatar
      name={member?.name || member?.email}
      alt={member?.email ? `${member.email} avatar` : undefined}
      size="sm"
    />
    <span className="flex min-w-0 flex-col gap-0.5">
      <Text lines={1}>{member?.email ?? 'Unknown member'}</Text>
      {member?.name ? (
        <Text variant="caption" color="secondary" lines={1}>
          {member.name}
        </Text>
      ) : null}
      {member?.id ? (
        <ID value={member.id} label="Copy member ID" truncate />
      ) : null}
    </span>
  </span>
)

const Joined = ({ member }: { member: TOrgMember }) =>
  member?.invite_id ? (
    <span className="flex min-w-0 flex-col gap-0.5">
      <Text variant="caption">Invited</Text>
      <Time value={member?.created_at} format="relative" />
    </span>
  ) : (
    <Time value={member?.joined_at} format="relative" />
  )

const MemberActions = ({
  member,
  currentAccountId,
  admin,
  onChangeRole,
  onRemove,
  onResend,
  onRevoke,
}: {
  member: TOrgMember
  currentAccountId?: string
  admin: boolean
  onChangeRole: (member: TOrgMember) => void
  onRemove: (member: TOrgMember) => void
  onResend: (member: TOrgMember) => void
  onRevoke: (member: TOrgMember) => void
}) => (
  <Dropdown
    align="end"
    trigger={
      <Button
        size="sm"
        variant="ghost"
        iconOnly
        aria-label={`${member?.email ?? 'Member'} actions`}
        tooltip="Team member actions"
      >
        <Icon variant="DotsThreeIcon" size={16} aria-hidden />
      </Button>
    }
  >
    <Menu>
      {member?.account_id ? (
        <>
          {admin && member.account_id !== currentAccountId ? (
            <MenuItem onSelect={() => onChangeRole(member)}>
              Change role
            </MenuItem>
          ) : null}
          <MenuItem tone="danger" onSelect={() => onRemove(member)}>
            Remove from org
          </MenuItem>
        </>
      ) : null}
      {member?.invite_id ? (
        <>
          <MenuItem onSelect={() => onResend(member)}>Resend invite</MenuItem>
          {admin ? (
            <MenuItem tone="danger" onSelect={() => onRevoke(member)}>
              Revoke invite
            </MenuItem>
          ) : null}
        </>
      ) : null}
    </Menu>
  </Dropdown>
)

export const columnsFor = ({
  roles,
  currentAccountId,
  admin,
  onChangeRole,
  onRemove,
  onResend,
  onRevoke,
}: Pick<
  ITeamTable,
  | 'roles'
  | 'currentAccountId'
  | 'admin'
  | 'onChangeRole'
  | 'onRemove'
  | 'onResend'
  | 'onRevoke'
>): ColumnDef<TOrgMember>[] => [
  {
    id: 'member',
    header: 'Member',
    size: 280,
    cell: ({ row }) => <MemberIdentity member={row.original} />,
  },
  {
    id: 'role',
    header: 'Role',
    size: 160,
    cell: ({ row }) => <Text lines={1}>{roleTitle(row.original, roles)}</Text>,
  },
  {
    id: 'status',
    header: 'Status',
    size: 120,
    cell: ({ row }) => <Status status={row.original?.status ?? 'unknown'} />,
  },
  {
    id: 'joined',
    header: 'Joined',
    size: 130,
    cell: ({ row }) => <Joined member={row.original} />,
  },
  {
    id: 'actions',
    header: '',
    size: 56,
    cell: ({ row }) => (
      <MemberActions
        member={row.original}
        currentAccountId={currentAccountId}
        admin={admin}
        onChangeRole={onChangeRole}
        onRemove={onRemove}
        onResend={onResend}
        onRevoke={onRevoke}
      />
    ),
  },
]

const TeamCard = ({
  member,
  roles,
  currentAccountId,
  admin,
  onChangeRole,
  onRemove,
  onResend,
  onRevoke,
}: {
  member: TOrgMember
} & Pick<
  ITeamTable,
  | 'roles'
  | 'currentAccountId'
  | 'admin'
  | 'onChangeRole'
  | 'onRemove'
  | 'onResend'
  | 'onRevoke'
>) => (
  <Card className="flex h-full flex-col gap-4">
    <span className="flex min-w-0 items-start justify-between gap-3">
      <MemberIdentity member={member} />
      <MemberActions
        member={member}
        currentAccountId={currentAccountId}
        admin={admin}
        onChangeRole={onChangeRole}
        onRemove={onRemove}
        onResend={onResend}
        onRevoke={onRevoke}
      />
    </span>
    <dl className="grid grid-cols-3 gap-3">
      <div className="min-w-0">
        <Text as="dt" variant="label" color="tertiary">
          Role
        </Text>
        <Text as="dd" className="mt-0.5" lines={1}>
          {roleTitle(member, roles)}
        </Text>
      </div>
      <div className="min-w-0">
        <Text as="dt" variant="label" color="tertiary">
          Status
        </Text>
        <Text as="dd" className="mt-0.5">
          <Status status={member?.status ?? 'unknown'} />
        </Text>
      </div>
      <div className="min-w-0">
        <Text as="dt" variant="label" color="tertiary">
          Joined
        </Text>
        <Text as="dd" className="mt-0.5">
          <Joined member={member} />
        </Text>
      </div>
    </dl>
  </Card>
)

const filterControl = (filter: ITeamFilter) => (
  <FilterDropdown
    key={filter.label}
    label={filter.label}
    options={filter.options}
    selected={filter.selected}
    onToggle={filter.onToggle}
    onIsolate={filter.onIsolate}
    onReset={filter.onReset}
    constrained={filter.constrained}
  />
)

export const TeamTable = ({
  members,
  roles,
  currentAccountId,
  admin,
  search,
  onSearchChange,
  offset,
  pageSize,
  hasNext,
  onOffsetChange,
  statusFilter,
  roleFilter,
  onChangeRole,
  onRemove,
  onResend,
  onRevoke,
  loading = false,
  fetching = false,
  error,
}: ITeamTable) => (
  <div className="flex min-w-0 flex-col gap-4">
    <Table
      data={members}
      columns={columnsFor({
        roles,
        currentAccountId,
        admin,
        onChangeRole,
        onRemove,
        onResend,
        onRevoke,
      })}
      getRowId={(member) => member?.id ?? ''}
      loading={loading}
      loadingLabel="Loading team members"
      emptyState={error ? 'Team members failed to load' : 'No team members yet'}
      toolbar={
        <>
          <ListSearch
            value={search}
            onValueChange={onSearchChange}
            placeholder="Search by email or name"
            aria-label="Search team members"
            className="w-full max-w-sm"
          />
          {filterControl(statusFilter)}
          {roleFilter ? filterControl(roleFilter) : null}
        </>
      }
      renderCard={({ row }) => (
        <TeamCard
          member={row.original}
          roles={roles}
          currentAccountId={currentAccountId}
          admin={admin}
          onChangeRole={onChangeRole}
          onRemove={onRemove}
          onResend={onResend}
          onRevoke={onRevoke}
        />
      )}
    />
    <Pagination
      label="Team members pagination"
      offset={offset}
      pageSize={pageSize}
      hasNext={hasNext}
      loading={fetching}
      onOffsetChange={onOffsetChange}
    />
  </div>
)
