import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Button } from '@/components/common/Button'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TStaticToken } from '@/types'

export const LegacyPersonalTokens = ({
  data,
  isLoading,
  onRevoke,
}: {
  data: TStaticToken[]
  isLoading: boolean
  onRevoke: (token: TStaticToken) => void
}) => {
  const columns: ColumnDef<TStaticToken>[] = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <Text variant="body" weight="strong">
            {props.getValue<string>() || 'Unnamed token'}
          </Text>
        ),
      },
      {
        header: 'Created',
        accessorKey: 'created_at',
        cell: (props) => (
          <Time time={props.getValue<string>()} format="relative" />
        ),
      },
      {
        header: 'Expires',
        accessorKey: 'expires_at',
        cell: (props) => (
          <Time time={props.getValue<string>()} format="short-datetime" />
        ),
      },
      {
        id: 'action',
        header: 'Action',
        cell: (props) => (
          <Button
            size="sm"
            variant="ghost"
            className="!text-red-800 dark:!text-red-500"
            onClick={() => onRevoke(props.row.original)}
          >
            Revoke
          </Button>
        ),
      },
    ],
    [onRevoke]
  )

  if (!isLoading && data.length === 0) return null

  return (
    <PageSection>
      <SectionHeader
        title="Legacy personal tokens"
        description="Personal tokens issued against your own account. Prefer service account tokens above — they keep working if you lose access."
      />
      <Table<TStaticToken>
        columns={columns}
        data={data}
        isLoading={isLoading}
        enableSearch={false}
        emptyStateProps={{
          emptyTitle: 'No legacy personal tokens',
          emptyMessage: 'Personal tokens created before service account tokens will appear here.',
        }}
      />
    </PageSection>
  )
}
