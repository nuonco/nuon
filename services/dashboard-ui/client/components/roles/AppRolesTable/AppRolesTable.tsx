import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel } from '@/components/surfaces/Panel'
import { AppRoleDetail } from '../AppRoleDetail'

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
  }[]
  permissions_boundary?: string
}

export const AppRolesTable = ({
  roles,
  isLoading,
}: {
  roles: TAppRole[]
  isLoading?: boolean
}) => {
  const columns = useMemo<ColumnDef<TAppRole, unknown>[]>(
    () => [
      {
        accessorKey: 'display_name',
        header: 'Name',
        cell: (info) => (
          <Text weight="strong">{info.getValue() as string}</Text>
        ),
      },
      {
        accessorKey: 'type',
        header: 'Type',
        cell: (info) => (
          <Badge variant="code" size="sm">
            {info.getValue() as string}
          </Badge>
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
            panelKey={row.original.id}
            childrenClassName="mt-12"
            heading={
              <div className="flex flex-col">
                <Text variant="h3">{row.original.display_name}</Text>
                <Text variant="subtext" theme="neutral" weight="normal">
                  {row.original.description}
                </Text>
              </div>
            }
            triggerButton={{
              size: 'sm',
              variant: 'ghost',
              children: 'View details',
            }}
          >
            <AppRoleDetail role={row.original} />
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
      enableSearch={false}
      emptyMessage="No roles found"
    />
  )
}
