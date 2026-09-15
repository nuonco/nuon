import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel } from '@/components/surfaces/Panel'
import { AppRoleDetail } from '../AppRoleDetail'
import type { TNamedIAMPolicy } from '@/lib/ctl-api/installs/get-install-app-permissions-config'

type TAppRole = {
  id?: string
  display_name?: string
  description?: string
  name?: string
  type?: string
  created_at?: string
  policies?: {
    id?: string
    name?: string
    managed_policy_name?: string
    contents?: string
    gcp_predefined_role?: string
    gcp_permissions?: string[]
    azure_built_in_roles?: string[]
    azure_actions?: string[]
  }[]
  named_policy_names?: string[]
  permissions_boundary?: string
}

const panelLinkClass =
  '!p-0 !h-auto !border-none !rounded-none !bg-transparent hover:!bg-transparent active:!bg-transparent focus:!shadow-none text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600'

const RolePanelHeading = ({ role }: { role: TAppRole }) => (
  <div className="flex flex-col">
    <Text variant="h3">{role.display_name}</Text>
    <Text variant="subtext" theme="neutral" weight="normal">
      {role.description}
    </Text>
  </div>
)

export const AppRolesTable = ({
  roles,
  namedPolicies = [],
  isLoading,
}: {
  roles: TAppRole[]
  namedPolicies?: TNamedIAMPolicy[]
  isLoading?: boolean
}) => {
  const columns = useMemo<ColumnDef<TAppRole, unknown>[]>(
    () => [
      {
        accessorKey: 'display_name',
        header: 'Role Name',
        cell: ({ row }) => (
          <Panel
            size="3/4"
            panelKey={`${row.original.id}-name`}
            heading={<RolePanelHeading role={row.original} />}
            triggerButton={{
              variant: 'ghost',
              className: panelLinkClass,
              children: row.original.display_name,
            }}
          >
            <AppRoleDetail role={row.original} namedPolicies={namedPolicies} />
          </Panel>
        ),
      },
      {
        accessorKey: 'type',
        header: 'Type',
        cell: (info) => (
          <Badge variant="code" theme="neutral">
            {info.getValue() as string}
          </Badge>
        ),
      },
      {
        id: 'named_policies',
        header: 'Named policies',
        cell: ({ row }) => (
          <Text variant="subtext">
            {row.original.named_policy_names?.length ?? 0}
          </Text>
        ),
      },
      {
        accessorKey: 'created_at',
        header: 'Created',
        cell: (info) => (
          <Time
            variant="subtext"
            time={info.getValue() as string}
            format="relative"
          />
        ),
      },
      {
        id: 'actions',
        header: '',
        enableSorting: false,
        cell: ({ row }) => (
          <Panel
            size="3/4"
            panelKey={`${row.original.id}-action`}
            heading={<RolePanelHeading role={row.original} />}
            triggerButton={{
              variant: 'ghost',
              className: panelLinkClass,
              children: (
                <span className="flex items-center gap-1.5">View role</span>
              ),
            }}
          >
            <AppRoleDetail role={row.original} namedPolicies={namedPolicies} />
          </Panel>
        ),
      },
    ],
    [namedPolicies]
  )

  return (
    <Table
      columns={columns}
      data={roles}
      isLoading={isLoading}
      enableSearch={false}
      emptyMessage="No roles found"
    />
  )
}
