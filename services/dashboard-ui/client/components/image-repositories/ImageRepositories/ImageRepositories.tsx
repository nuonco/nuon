import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Link } from '@/components/common/Link'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { ListPage } from '@/components/layout/ListPage'
import type { TPaginationMeta } from '@/lib/api'
import type { TAPIError, TImageRepository } from '@/types'

export const ImageRepositories = ({
  orgId,
  repositories,
  isLoading = false,
  error,
  pagination,
}: {
  orgId: string
  repositories: TImageRepository[]
  isLoading?: boolean
  error?: TAPIError | null
  pagination?: TPaginationMeta
}) => {
  const [params, setParams] = useSearchParams()
  const platform = params.get('platform') || undefined
  const filtered = !!(platform || params.get('q'))
  const basePath = `/${orgId}/settings/image-repositories`
  const columns = useMemo<ColumnDef<TImageRepository>[]>(
    () => [
      {
        accessorKey: 'image_url_prefix',
        header: 'Repository',
        cell: ({ row }) => (
          <Link href={`${basePath}/${row.original.id}`}>
            <Text family="mono" className="break-all">
              {row.original.image_url_prefix}
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
            displayVariant="name"
            colorVariant="color"
          />
        ),
      },
      {
        accessorKey: 'cloud_connection.name',
        header: 'Pull identity',
        cell: ({ row }) => (
          <div className="flex min-w-0 flex-col gap-1">
            <Text family="mono" className="break-all">
              {row.original.cloud_connection?.name || 'Unavailable identity'}
            </Text>
            <Text
              family="mono"
              variant="subtext"
              theme="neutral"
              className="break-all"
            >
              {row.original.cloud_connection?.principal}
            </Text>
          </div>
        ),
      },
      {
        accessorKey: 'cloud_connection.status',
        header: 'Authentication',
        cell: ({ row }) =>
          row.original.cloud_connection?.id ? (
            <Status
              status={row.original.cloud_connection.status}
              variant="badge"
            />
          ) : (
            <Status status="error">Unavailable</Status>
          ),
      },
      {
        accessorKey: 'created_at',
        header: 'Added',
        cell: ({ row }) => <Time time={row.original.created_at} />,
      },
    ],
    [basePath]
  )
  return (
    <ListPage
      title="Image repositories"
      description="Choose where Nuon pulls private images and which identity it uses for each repository."
      createAction={
        <Button variant="primary" href={`${basePath}/add`}>
          Add image repository
        </Button>
      }
    >
      <FormErrorBanner
        error={error}
        fallback="Image repositories failed to load"
      />
      <Table
        columns={columns}
        data={repositories}
        isLoading={isLoading}
        pagination={pagination}
        enableSearch={repositories.length > 0 || filtered || isLoading}
        searchPlaceholder="Search repositories..."
        filterActions={
          <RadioFilterDropdown
            id="image-repositories-cloud"
            label="Cloud"
            selected={platform}
            options={[
              { value: 'aws', label: 'AWS' },
              { value: 'gcp', label: 'Google Cloud' },
              { value: 'azure', label: 'Azure' },
            ]}
            onChange={(value) =>
              setParams((previous) => {
                const next = new URLSearchParams(previous)
                if (value) next.set('platform', value)
                else next.delete('platform')
                next.delete('offset')
                return next
              })
            }
          />
        }
        emptyStateProps={{
          emptyTitle: filtered
            ? 'No matching image repositories'
            : 'No image repositories yet',
          emptyMessage: filtered
            ? 'Change the cloud filter or try another search.'
            : 'Add a private repository, choose an existing pull identity or create one, then follow its setup instructions.',
          ...(!filtered
            ? {
                action: (
                  <Button href={`${basePath}/add`}>Add image repository</Button>
                ),
              }
            : {}),
        }}
      />
      <Text variant="subtext" theme="neutral">
        Authentication verifies the identity, not repository permissions. Image
        builds check actual pull access.
      </Text>
    </ListPage>
  )
}
