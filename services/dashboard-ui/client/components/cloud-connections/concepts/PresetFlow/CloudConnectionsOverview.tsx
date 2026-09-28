import { useMemo, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { ListPage } from '@/components/layout/ListPage'
import { ConnectionStatus } from './CloudConnectionDetailPage'
import { PresetFlow } from './PresetFlow'
import type { TCloudConnectionOverview } from './overviewMockData'

interface ICloudConnectionsOverview {
  connections: TCloudConnectionOverview[]
  isLoading?: boolean
}

export const CloudConnectionsOverview = ({
  connections,
  isLoading = false,
}: ICloudConnectionsOverview) => {
  const [isCreating, setIsCreating] = useState(false)

  const columns = useMemo<ColumnDef<TCloudConnectionOverview>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'Name',
        cell: ({ row }) => (
          <div className="flex min-w-0 max-w-lg flex-col gap-1">
            <Link
              href={`/org-mock-001/cloud-connections/${row.original.id}`}
            >
              <Text weight="strong">{row.original.name}</Text>
            </Link>
            <ClickToCopy className="max-w-full">
              <Text
                family="mono"
                variant="subtext"
                theme="neutral"
                className="min-w-0 max-w-96 overflow-hidden text-ellipsis"
                nowrap
              >
                {row.original.roleArn}
              </Text>
            </ClickToCopy>
          </div>
        ),
      },
      {
        accessorKey: 'targetId',
        header: 'Account',
        cell: ({ row }) => (
          <div className="flex flex-col gap-1">
            <Text family="mono" variant="subtext">
              {row.original.targetId}
            </Text>
            <Text family="mono" variant="subtext" theme="neutral">
              {row.original.region}
            </Text>
          </div>
        ),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => <ConnectionStatus connection={row.original} />,
      },
    ],
    []
  )

  if (isCreating) return <PresetFlow />

  const createButton = (
    <Button variant="primary" onClick={() => setIsCreating(true)}>
      Create cloud connection
    </Button>
  )

  return (
    <ListPage
      variant="page"
      title="Cloud connections"
      description="Manage the AWS roles Nuon can assume to operate installs in this org."
      createAction={createButton}
    >
      <Table<TCloudConnectionOverview>
        columns={columns}
        data={connections}
        isLoading={isLoading}
        enableSearch={connections.length > 0 || isLoading}
        searchPlaceholder="Search by name, account, or role ARN..."
        rowClassName={() =>
          'transition-colors hover:bg-black/5 dark:hover:bg-white/5'
        }
        emptyStateProps={{
          emptyTitle: 'No cloud connections yet',
          emptyMessage:
            'A cloud connection lets Nuon assume a role you control to manage resources in an AWS account. Create a connection to use it when creating an install.',
          action: (
            <Button variant="secondary" onClick={() => setIsCreating(true)}>
              Create cloud connection
            </Button>
          ),
        }}
      />
    </ListPage>
  )
}
