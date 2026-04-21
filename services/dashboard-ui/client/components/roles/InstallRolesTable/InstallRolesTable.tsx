import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { type IPagination } from '@/components/common/Pagination'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel } from '@/components/surfaces/Panel'
import { InstallRoleDetail } from '../InstallRoleDetail'
import type { TInstallRole } from '@/lib/ctl-api/installs/get-latest-install-roles'

const panelLinkClass =
  '!p-0 !h-auto !border-none !rounded-none !bg-transparent hover:!bg-transparent active:!bg-transparent focus:!shadow-none text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600'

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
          <Panel
            size="3/4"
            panelKey={`${row.original.id}-name`}
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
              variant: 'ghost',
              className: panelLinkClass,
              children: (
                <span className="flex flex-col items-start">
                  {row.original.app_role_config?.display_name}
                  <ID>{row.original.id}</ID>
                </span>
              ),
            }}
          >
            <InstallRoleDetail installRole={row.original} />
          </Panel>
        ),
      },
      {
        accessorKey: 'app_role_config.type',
        header: 'Type',
        cell: ({ row }) => (
          <Badge variant="code" theme="neutral">
            {row.original.app_role_config?.type}
          </Badge>
        ),
      },
      {
        accessorKey: 'provisioned',
        header: 'Status',
        cell: ({ row }) => (
          <Status status={row.original.provisioned ? 'active' : 'inactive'}>
            {row.original.provisioned ? 'Provisioned' : 'Not provisioned'}
          </Status>
        ),
      },
      {
        accessorKey: 'last_used_at',
        header: 'Last used',
        cell: ({ row }) =>
          row.original.last_used_at ? (
            <Time
              variant="subtext"
              time={row.original.last_used_at}
              format="relative"
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              Never
            </Text>
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
              variant: 'ghost',
              className: panelLinkClass,
              children: (
                <span className="flex items-center gap-1.5">
                  View <Icon variant="CaretRightIcon" />
                </span>
              ),
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
