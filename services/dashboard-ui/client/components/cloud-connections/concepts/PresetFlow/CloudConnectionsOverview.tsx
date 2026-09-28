import { useMemo, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { StatusWithDescription } from '@/components/common/StatusWithDescription'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ListPage } from '@/components/layout/ListPage'
import { PresetFlow } from './PresetFlow'
import type { TCloudConnectionOverview } from './overviewMockData'

const RowActions = ({
  connection,
}: {
  connection: TCloudConnectionOverview
}) => {
  const isInUse = connection.installCount > 0
  return (
    <Dropdown
      id={`cloud-connection-actions-${connection.id}`}
      buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
      hideIcon
      variant="ghost"
      buttonClassName="!p-1"
      alignment="right"
    >
      <Menu>
        <Button onClick={() => {}}>Verify now</Button>
        <Button onClick={() => {}}>View setup</Button>
        <hr />
        <Button
          variant="danger"
          disabled={isInUse}
          tooltipProps={
            isInUse
              ? {
                  tipContent: `Cannot delete — ${connection.installCount} install${connection.installCount === 1 ? '' : 's'} use this connection`,
                }
              : undefined
          }
          onClick={() => {}}
        >
          Delete connection
        </Button>
      </Menu>
    </Dropdown>
  )
}

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
        cell: ({ row }) => <Text weight="strong">{row.original.name}</Text>,
      },
      {
        accessorKey: 'targetId',
        header: 'AWS account',
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
        accessorKey: 'roleArn',
        header: 'Role ARN',
        enableSorting: false,
        cell: ({ row }) => (
          <ClickToCopy className="max-w-56">
            <Text
              family="mono"
              variant="subtext"
              theme="neutral"
              className="block truncate"
            >
              {row.original.roleArn}
            </Text>
          </ClickToCopy>
        ),
      },
      {
        accessorKey: 'preset',
        header: 'Preset',
        cell: ({ row }) => <Badge>{row.original.preset}</Badge>,
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => (
          <StatusWithDescription
            statusProps={{ status: row.original.status, variant: 'badge' }}
            tooltipProps={
              row.original.statusMessage
                ? { tipContent: row.original.statusMessage }
                : undefined
            }
          />
        ),
      },
      {
        accessorKey: 'lastVerifiedAt',
        header: 'Last verified',
        cell: ({ row }) =>
          row.original.lastVerifiedAt ? (
            <Time
              time={row.original.lastVerifiedAt}
              format="relative"
              variant="subtext"
              className="whitespace-nowrap"
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              Never
            </Text>
          ),
      },
      {
        accessorKey: 'installCount',
        header: 'Used by',
        cell: ({ row }) => (
          <Text variant="subtext">
            {row.original.installCount} install
            {row.original.installCount === 1 ? '' : 's'}
          </Text>
        ),
      },
      {
        id: 'actions',
        header: '',
        enableSorting: false,
        cell: ({ row }) => <RowActions connection={row.original} />,
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
