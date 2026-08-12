import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import type { TRoleInfo } from '@/types'
import { DeleteRoleButton } from '@/components/access-roles/DeleteRole'
import { EditRoleButton } from '@/components/access-roles/RoleForm'
import { entriesFromRole, entrySummary } from '../permissions'

const CONTEXT_LABELS: Record<string, string> = {
  team: 'Team',
  service_account: 'Service accounts',
  api_token: 'API tokens',
  oidc_trust_policy: 'OIDC',
}

const ActionCell = ({ role }: { role: TRoleInfo }) => {
  if (role.managed) {
    return (
      <Text variant="subtext" theme="neutral">
        —
      </Text>
    )
  }

  return (
    <Dropdown
      id={`action-${role.id}`}
      buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
      hideIcon
      variant="ghost"
      buttonClassName="!p-1"
      alignment="right"
    >
      <Menu>
        <span>
          <EditRoleButton role={role} isMenuButton />
        </span>
        <span>
          <DeleteRoleButton role={role} isMenuButton />
        </span>
      </Menu>
    </Dropdown>
  )
}

export const AccessRolesTable = ({
  data,
  isLoading,
  nameFor,
}: {
  data: TRoleInfo[]
  isLoading: boolean
  nameFor?: (id: string) => string | undefined
}) => {
  const columns: ColumnDef<TRoleInfo>[] = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'title',
        cell: (props) => (
          <div className="flex items-center gap-2">
            <Text variant="body" weight="strong">
              {props.getValue<string>() || props.row.original.role_type}
            </Text>
            {props.row.original.managed ? (
              <Badge theme="neutral" size="sm">
                Built in
              </Badge>
            ) : null}
          </div>
        ),
      },
      {
        header: 'Description',
        accessorKey: 'description',
        cell: (props) => (
          <Text variant="subtext" theme="neutral">
            {props.getValue<string>() || '—'}
          </Text>
        ),
      },
      {
        id: 'permissions',
        header: 'Permissions',
        cell: (props) => {
          const entries = entriesFromRole(props.row.original.policies)

          if (props.row.original.managed) {
            return (
              <Text variant="subtext" theme="neutral">
                Org-wide
              </Text>
            )
          }
          if (entries.length === 0) {
            return (
              <Text variant="subtext" theme="neutral">
                None
              </Text>
            )
          }

          return (
            <div className="flex flex-col gap-0.5">
              {entries.slice(0, 3).map((entry, i) => (
                <Text key={i} variant="subtext" theme="neutral">
                  {entrySummary(entry, nameFor)}
                </Text>
              ))}
              {entries.length > 3 ? (
                <Text variant="subtext" theme="neutral">
                  and {entries.length - 3} more
                </Text>
              ) : null}
            </div>
          )
        },
      },
      {
        id: 'applies_to',
        header: 'Assignable to',
        cell: (props) => {
          const contexts = props.row.original.applies_to ?? []

          if (contexts.length === 0) {
            return (
              <Text variant="subtext" theme="warn">
                Nowhere
              </Text>
            )
          }

          return (
            <div className="flex flex-wrap gap-1">
              {contexts.map((context) => (
                <Badge key={context} theme="default" size="sm">
                  {CONTEXT_LABELS[context] ?? context}
                </Badge>
              ))}
            </div>
          )
        },
      },
      {
        id: 'action',
        header: 'Action',
        cell: (props) => <ActionCell role={props.row.original} />,
      },
    ],
    [nameFor]
  )

  if (isLoading) {
    return <AccessRolesTableSkeleton />
  }

  return (
    <Table<TRoleInfo>
      columns={columns}
      data={data}
      searchPlaceholder="Search roles"
      emptyStateProps={{
        emptyTitle: 'No roles yet',
        emptyMessage:
          'Create a role to grant scoped access to specific apps and installs.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TRoleInfo>[] = [
  { header: 'Name', accessorKey: 'title' },
  { header: 'Description', accessorKey: 'description' },
  { header: 'Permissions', id: 'permissions' },
  { header: 'Assignable to', id: 'applies_to' },
  { header: 'Action', id: 'action' },
]

export const AccessRolesTableSkeleton = () => (
  <TableSkeleton<TRoleInfo> columns={skeletonColumns} skeletonRows={5} />
)
