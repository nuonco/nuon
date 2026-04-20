import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { type IPagination } from '@/components/common/Pagination'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel } from '@/components/surfaces/Panel'
import { InstallRoleDetail } from '../InstallRoleDetail'
import type { TInstallRole } from '@/lib/ctl-api/installs/get-latest-install-roles'

export const InstallRolesTable = ({
  roles,
  isLoading,
  pagination,
}: {
  roles: TInstallRole[]
  isLoading?: boolean
  pagination?: Omit<IPagination, 'position'>
}) => {
  const columns = useMemo<ColumnDef<TInstallRole, unknown>[]>(
    () => [
      {
        accessorKey: 'app_role_config.display_name',
        header: 'Name',
        cell: ({ row }) => (
          <Text weight="strong">
            {row.original.app_role_config?.display_name}
          </Text>
        ),
      },
      {
        accessorKey: 'app_role_config.type',
        header: 'Type',
        cell: ({ row }) => (
          <Badge variant="code" size="sm">
            {row.original.app_role_config?.type}
          </Badge>
        ),
      },
      {
        accessorKey: 'app_role_config.created_at',
        header: 'Created',
        cell: ({ row }) => (
          <Time
            variant="subtext"
            time={row.original.app_role_config?.created_at}
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
            panelKey={row.original.id}

            heading={
              <div className="flex flex-col">
                <Text variant="h3">
                  {row.original.app_role_config?.display_name}
                </Text>
                <Text variant="subtext" theme="neutral" weight="normal">
                  {row.original.app_role_config?.description}
                </Text>
              </div>
            }
            triggerButton={{
              size: 'sm',
              variant: 'ghost',
              children: 'View details',
            }}
          >
            <InstallRoleDetail installRole={row.original} />
          </Panel>
        ),
      },
    ],
    []
  )

  return (
    <Table
      columns={columns}
      data={roles}
      isLoading={isLoading}
      pagination={pagination}
      searchPlaceholder="Search role name..."
      emptyMessage="No roles found"
    />
  )
}
