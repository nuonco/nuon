import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { ListPage } from '@/components/layout/ListPage'
import type { TAPIError, TCloudConnectionSummary } from '@/types'
import type { TPaginationMeta } from '@/lib/api'
import { ConnectionStatus } from '../ConnectionStatus'

export const ConnectionsList = ({
  connections,
  orgId,
  isLoading = false,
  error,
  pagination,
}: {
  connections: TCloudConnectionSummary[]
  orgId: string
  isLoading?: boolean
  error?: TAPIError | null
  pagination?: TPaginationMeta & { offset?: number }
}) => {
  const createHref = `/${orgId}/settings/cloud-connections/create`
  const [params] = useSearchParams()
  const search = (params.get('q') || '').toLowerCase()
  const columns = useMemo<ColumnDef<TCloudConnectionSummary>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'Name',
        cell: ({ row }) => (
          <div className="flex min-w-0 max-w-lg flex-col gap-1">
            <Link
              href={`/${orgId}/settings/cloud-connections/${row.original.id}`}
            >
              <Text family="mono" weight="strong">
                {row.original.name}
              </Text>
            </Link>
            <ClickToCopy>
              <Text
                family="mono"
                variant="subtext"
                theme="neutral"
                className="max-w-96 truncate"
              >
                {row.original.principal}
              </Text>
            </ClickToCopy>
          </div>
        ),
      },
      {
        accessorKey: 'target_id',
        header: 'Account',
        cell: ({ row }) => (
          <Text family="mono" variant="subtext">
            {row.original.target_id}
            {row.original.default_region ? (
              <Text as="span" theme="neutral" family="mono" variant="subtext">
                {' '}
                · {row.original.default_region}
              </Text>
            ) : null}
          </Text>
        ),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => <ConnectionStatus connection={row.original} />,
      },
    ],
    [orgId]
  )
  return (
    <ListPage
      title="Cloud connections"
      description="Manage the AWS roles Nuon can assume to operate installs in this org."
      createAction={
        <Button variant="primary" href={createHref}>
          Create cloud connection
        </Button>
      }
    >
      <FormErrorBanner
        error={error}
        fallback="Cloud connections failed to load"
      />
      <Table
        columns={columns}
        data={connections}
        isLoading={isLoading}
        pagination={pagination}
        enableSearch={connections.length > 0 || isLoading || !!search}
        searchPlaceholder="Search by name, account, or role ARN..."
        emptyStateProps={
          search
            ? {
                emptyTitle: 'No cloud connections found',
                emptyMessage: 'Try another name, account, or role ARN.',
              }
            : {
                emptyTitle: 'No cloud connections yet',
                emptyMessage:
                  'A cloud connection lets Nuon assume a role you control to manage resources in an AWS account. Create a connection to use it when creating an install.',
                action: (
                  <Button variant="secondary" href={createHref}>
                    Create cloud connection
                  </Button>
                ),
              }
        }
      />
    </ListPage>
  )
}
