import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponentType } from '@/types'
import type { TDotNode, TDotEdge } from './parse-dot'

type TGraphTableRow = {
  id: string
  label: string
  type: string
  dependsOn: string[]
  requiredBy: string[]
}

export interface IComponentsGraphTable {
  nodes: TDotNode[]
  edges: TDotEdge[]
  isLoading?: boolean
}

const columns: ColumnDef<TGraphTableRow, any>[] = [
  {
    header: 'Component',
    accessorKey: 'label',
    cell: ({ row }) => (
      <span className="flex items-center gap-2">
        <ComponentType
          type={row.original.type as TComponentType}
          displayVariant="icon-only"
          variant="subtext"
        />
        <Text variant="body">{row.original.label}</Text>
      </span>
    ),
  },
  {
    header: 'Depends on',
    accessorKey: 'dependsOn',
    enableSorting: false,
    cell: ({ row }) => (
      <Text variant="subtext" theme="neutral">
        {row.original.dependsOn.length
          ? row.original.dependsOn.join(', ')
          : '—'}
      </Text>
    ),
  },
  {
    header: 'Required by',
    accessorKey: 'requiredBy',
    enableSorting: false,
    cell: ({ row }) => (
      <Text variant="subtext" theme="neutral">
        {row.original.requiredBy.length
          ? row.original.requiredBy.join(', ')
          : '—'}
      </Text>
    ),
  },
]

export const ComponentsGraphTable = ({
  nodes,
  edges,
  isLoading,
}: IComponentsGraphTable) => {
  const rows = useMemo<TGraphTableRow[]>(() => {
    const labelById = new Map(nodes.map((n) => [n.id, n.label]))
    return nodes
      .map((node) => ({
        id: node.id,
        label: node.label,
        type: node.type,
        dependsOn: edges
          .filter((e) => e.target === node.id)
          .map((e) => labelById.get(e.source) ?? e.source)
          .sort(),
        requiredBy: edges
          .filter((e) => e.source === node.id)
          .map((e) => labelById.get(e.target) ?? e.target)
          .sort(),
      }))
      .sort((a, b) => a.label.localeCompare(b.label))
  }, [nodes, edges])

  return (
    <Table<TGraphTableRow>
      columns={columns}
      data={rows}
      isLoading={isLoading}
      enableSearch={false}
      enableSorting={false}
      emptyStateProps={{
        emptyTitle: 'No components yet',
        emptyMessage:
          'Components will appear here once this config version includes them.',
      }}
    />
  )
}
