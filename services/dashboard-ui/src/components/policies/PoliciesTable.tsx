'use client'

import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import type { TAppPolicyConfig } from '@/types'

type TPolicyRow = {
  id: string
  name: string
  type: string
  engine: string
  components: string[]
  contents: string
  createdAt: string
}

function extractPolicyName(contents: string, engine: string): string {
  if (!contents) return 'Unnamed Policy'

  if (engine === 'kyverno') {
    const match = contents.match(/metadata:\s*\n\s*name:\s*["']?([^"'\n]+)["']?/)
    if (match) return match[1].trim()
  }

  if (engine === 'opa') {
    const match = contents.match(/package\s+([^\s]+)/)
    if (match) return match[1].trim()
  }

  return 'Unnamed Policy'
}

function formatPolicyType(type: string): string {
  return type
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function parsePolicyToTableData(policies: TAppPolicyConfig[]): TPolicyRow[] {
  return policies.map((policy) => ({
    id: policy.id || '',
    name:
      policy.name || extractPolicyName(policy.contents || '', policy.engine || ''),
    type: policy.type || '',
    engine: policy.engine || '',
    components: policy.components || [],
    contents: policy.contents || '',
    createdAt: policy.created_at || '',
  }))
}

export const policiesTableColumns: ColumnDef<TPolicyRow>[] = [
  {
    accessorKey: 'name',
    header: 'Policy Name',
    cell: (info) => (
      <span>
        <Text variant="body">{info.getValue() as string}</Text>
        <ID>{info.row.original.id}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'type',
    header: 'Type',
    cell: (info) => (
      <Badge theme="default" size="sm">
        {formatPolicyType(info.getValue() as string)}
      </Badge>
    ),
  },
  {
    accessorKey: 'engine',
    header: 'Engine',
    cell: (info) => (
      <Badge
        theme={info.getValue() === 'kyverno' ? 'brand' : 'info'}
        size="sm"
      >
        {(info.getValue() as string).toUpperCase()}
      </Badge>
    ),
  },
  {
    accessorKey: 'components',
    header: 'Components',
    cell: (info) => {
      const components = info.getValue() as string[]
      const isAllComponents =
        !components ||
        components.length === 0 ||
        (components.length === 1 && components[0] === '*')
      if (isAllComponents) {
        return (
          <Text variant="subtext" className="italic">
            All components
          </Text>
        )
      }
      return (
        <div className="flex flex-wrap gap-1">
          {components.slice(0, 3).map((comp) => (
            <Badge key={comp} theme="neutral" size="sm">
              {comp}
            </Badge>
          ))}
          {components.length > 3 && (
            <Text variant="subtext">+{components.length - 3} more</Text>
          )}
        </div>
      )
    },
  },
  {
    accessorKey: 'id',
    id: 'actions',
    header: '',
    enableSorting: false,
    cell: () => null,
  },
]

export const PoliciesTable = ({
  policies,
  orgId,
  appId,
}: {
  policies: TAppPolicyConfig[]
  orgId: string
  appId: string
}) => {
  const data = parsePolicyToTableData(policies)

  const columns: ColumnDef<TPolicyRow>[] = policiesTableColumns.map((col) => {
    if (col.id === 'actions') {
      return {
        ...col,
        cell: (info) => (
          <Button
            variant="ghost"
            size="sm"
            href={`/${orgId}/apps/${appId}/policies/${info.row.original.id}`}
          >
            View <Icon variant="CaretRightIcon" />
          </Button>
        ),
      }
    }
    return col
  })

  return (
    <Table<TPolicyRow>
      columns={columns}
      data={data}
      emptyStateProps={{
        emptyMessage: 'No policies configured for this app.',
        emptyTitle: 'No policies',
      }}
      pagination={{
        limit: data.length || 10,
        offset: 0,
        hasNext: false,
      }}
    />
  )
}
