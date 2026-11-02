import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Link } from '@/components/common/Link'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
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
  const [params, setParams] = useSearchParams()
  const search = params.get('q') || ''
  const platform = params.get('platform') || undefined
  const preset = params.get('preset') || undefined
  const isFiltered = !!(search || platform || preset)
  const setFilter = (key: string, value?: string) => {
    setParams((previous) => {
      const next = new URLSearchParams(previous)
      if (value) next.set(key, value)
      else next.delete(key)
      next.delete('offset')
      return next
    })
  }
  const columns = useMemo<ColumnDef<TCloudConnectionSummary>[]>(
    () => [
      {
        accessorKey: 'name',
        header: 'Name',
        cell: ({ row }) => (
          <Link
            href={`/${orgId}/settings/cloud-connections/${row.original.id}`}
          >
            <Text family="mono" weight="strong" className="break-all">
              {row.original.name}
            </Text>
          </Link>
        ),
      },
      {
        accessorKey: 'platform',
        header: 'Cloud',
        cell: ({ row }) => (
          <CloudPlatform
            platform={row.original.platform}
            displayVariant={row.original.platform === 'gcp' ? 'name' : 'abbr'}
            colorVariant="color"
            variant="subtext"
          />
        ),
      },
      {
        accessorKey: 'target_id',
        header: 'Cloud target',
        cell: ({ row }) => (
          <div className="flex flex-col gap-1">
            <Text family="mono" variant="subtext">
              {row.original.target_id}
            </Text>
            {row.original.default_region ? (
              <Text theme="neutral" family="mono" variant="subtext">
                {row.original.default_region}
              </Text>
            ) : null}
          </div>
        ),
      },
      {
        accessorKey: 'preset',
        header: 'Access',
        cell: ({ row }) => {
          const connection = row.original
          const isImage =
            connection.preset === 'ecr' ||
            connection.preset === 'gar' ||
            connection.preset === 'acr'
          return (
            <div className="flex flex-col gap-1">
              <Text variant="subtext">
                {isImage
                  ? 'Image repositories'
                  : connection.preset === 'stacks'
                    ? 'Install stacks'
                    : 'Custom'}
              </Text>
              {isImage && (
                <Text variant="subtext" theme="neutral">
                  {connection.preset === 'acr'
                    ? 'Azure ACR'
                    : connection.preset === 'gar'
                      ? 'Artifact Registry'
                      : 'Amazon ECR'}
                </Text>
              )}
            </div>
          )
        },
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
      description="Manage cloud identities Nuon can use for install stacks and image repository access in this org."
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
        enableSearch={connections.length > 0 || isLoading || isFiltered}
        searchPlaceholder="Search connections..."
        filterActions={
          <>
            <RadioFilterDropdown
              id="cloud-connections-filter-cloud"
              label="Cloud"
              options={[
                { value: 'aws', label: 'AWS' },
                { value: 'gcp', label: 'Google Cloud' },
                { value: 'azure', label: 'Azure' },
              ]}
              selected={platform}
              onChange={(value) => setFilter('platform', value)}
            />
            <RadioFilterDropdown
              id="cloud-connections-filter-access"
              label="Access"
              options={[
                { value: 'stacks', label: 'Install stacks' },
                { value: 'ecr,gar,acr', label: 'Image repositories' },
                { value: 'custom', label: 'Custom' },
              ]}
              selected={preset}
              onChange={(value) => setFilter('preset', value)}
            />
          </>
        }
        emptyStateProps={
          isFiltered
            ? {
                emptyTitle: 'No cloud connections found',
                emptyMessage:
                  'Change the cloud or access filters, or try another search.',
              }
            : {
                emptyTitle: 'No cloud connections yet',
                emptyMessage:
                  'A cloud connection gives Nuon access to cloud resources through an identity you control. Create a connection for install stacks or private image repositories.',
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
