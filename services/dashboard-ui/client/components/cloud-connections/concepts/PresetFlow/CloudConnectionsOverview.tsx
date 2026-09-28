import { useMemo, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { ListPage } from '@/components/layout/ListPage'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  CloudConnectionDetailsPanel,
  ConnectionStatus,
} from './CloudConnectionDetailsPanel'
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
  const { addPanel } = useSurfaces()
  const openDetails = (
    connection: TCloudConnectionOverview,
    initialSection: 'summary' | 'policy' = 'summary'
  ) =>
    addPanel(
      <CloudConnectionDetailsPanel
        connection={connection}
        initialSection={initialSection}
      />
    )

  const columns = useMemo<ColumnDef<TCloudConnectionOverview>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'Name',
        cell: ({ row }) => (
          <div className="flex min-w-0 max-w-lg flex-col gap-1">
            <Text weight="strong">{row.original.name}</Text>
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
      {
        id: 'actions',
        header: '',
        enableSorting: false,
        cell: ({ row }) => (
          <div className="flex items-center justify-end gap-1 whitespace-nowrap">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => openDetails(row.original, 'policy')}
            >
              View policy
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => openDetails(row.original)}
            >
              View details
            </Button>
          </div>
        ),
      },
    ],
    [addPanel]
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
