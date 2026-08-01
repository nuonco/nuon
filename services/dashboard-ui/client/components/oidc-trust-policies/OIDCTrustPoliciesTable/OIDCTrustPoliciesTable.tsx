import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Code } from '@/components/common/Code'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DeleteOIDCTrustPolicyButton } from '@/components/oidc-trust-policies/DeleteOIDCTrustPolicy'
import { EditOIDCTrustPolicyButton } from '@/components/oidc-trust-policies/EditOIDCTrustPolicy'
import type { TOIDCTrustPolicy } from '@/types'

export const OIDCTrustPoliciesTable = ({
  data,
  isLoading,
}: {
  data: TOIDCTrustPolicy[]
  isLoading: boolean
}) => {
  const columns: ColumnDef<TOIDCTrustPolicy>[] = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <Text variant="subtext" weight="strong">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Issuer URL',
        accessorKey: 'issuer_url',
        cell: (props) => (
          <Code variant="inline" className="!px-2 !py-1 w-fit">
            {props.getValue<string>()}
          </Code>
        ),
      },
      {
        header: 'Audience',
        accessorKey: 'audience',
        cell: (props) => (
          <Text variant="subtext" theme="neutral">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Role',
        accessorKey: 'role',
        cell: (props) => (
          <Text variant="subtext" theme="neutral">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Enabled',
        accessorKey: 'enabled',
        cell: (props) =>
          props.getValue<boolean>() ? (
            <Text variant="subtext">Enabled</Text>
          ) : (
            <Text variant="subtext" theme="neutral">
              Disabled
            </Text>
          ),
      },
      {
        header: 'Last used',
        accessorKey: 'last_used_at',
        cell: (props) => {
          const time = props.getValue<string | undefined>()
          return time ? (
            <Time variant="subtext" time={time} format="relative" />
          ) : (
            <Text variant="subtext" theme="neutral">
              Never
            </Text>
          )
        },
      },
      {
        id: 'action',
        header: '',
        cell: (props) => (
          <div className="flex justify-end gap-1">
            <EditOIDCTrustPolicyButton policy={props.row.original} size="sm" />
            <DeleteOIDCTrustPolicyButton
              policy={props.row.original}
              size="sm"
            />
          </div>
        ),
      },
    ],
    []
  )

  if (isLoading) {
    return <OIDCTrustPoliciesTableSkeleton />
  }

  return (
    <Table<TOIDCTrustPolicy>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No OIDC trust policies configured',
        emptyMessage:
          'Create a trust policy to let a CI/CD provider exchange OIDC tokens for org access.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TOIDCTrustPolicy>[] = [
  { header: 'Name', accessorKey: 'name' },
  { header: 'Issuer URL', accessorKey: 'issuer_url' },
  { header: 'Audience', accessorKey: 'audience' },
  { header: 'Role', accessorKey: 'role' },
  { header: 'Enabled', accessorKey: 'enabled' },
  { header: 'Last used', accessorKey: 'last_used_at' },
  { header: '', id: 'action' },
]

export const OIDCTrustPoliciesTableSkeleton = () => (
  <TableSkeleton<TOIDCTrustPolicy> columns={skeletonColumns} skeletonRows={3} />
)
