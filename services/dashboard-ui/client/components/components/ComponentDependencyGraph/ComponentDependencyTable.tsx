import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponentType, TTheme } from '@/types'
import type { GraphNode, GraphEdge } from './ComponentDependencyGraph'

type TRole = GraphNode['role']

type TDependencyRow = {
  id: string
  name: string
  type?: GraphNode['type']
  role: TRole
  dependsOn: string[]
  requiredBy: string[]
}

const ROLE_ORDER: Record<TRole, number> = {
  current: 0,
  dependency: 1,
  dependent: 2,
}

const ROLE_THEME: Record<TRole, TTheme> = {
  current: 'brand',
  dependency: 'info',
  dependent: 'neutral',
}

const ROLE_LABEL: Record<TRole, string> = {
  current: 'This component',
  dependency: 'Dependency',
  dependent: 'Dependent',
}

export interface IComponentDependencyTable {
  nodes: GraphNode[]
  edges: GraphEdge[]
  currentId: string
  basePath: string
  onNavigate?: () => void
}

export const ComponentDependencyTable = ({
  nodes,
  edges,
  currentId,
  basePath,
  onNavigate,
}: IComponentDependencyTable) => {
  const rows = useMemo<TDependencyRow[]>(() => {
    const nameById = new Map(nodes.map((n) => [n.id, n.name]))
    return nodes
      .map((node) => ({
        id: node.id,
        name: node.name,
        type: node.type,
        role: node.role,
        dependsOn: edges
          .filter((e) => e.targetId === node.id)
          .map((e) => nameById.get(e.sourceId) ?? e.sourceId)
          .sort(),
        requiredBy: edges
          .filter((e) => e.sourceId === node.id)
          .map((e) => nameById.get(e.targetId) ?? e.targetId)
          .sort(),
      }))
      .sort(
        (a, b) =>
          ROLE_ORDER[a.role] - ROLE_ORDER[b.role] ||
          a.name.localeCompare(b.name)
      )
  }, [nodes, edges])

  const columns = useMemo<ColumnDef<TDependencyRow, any>[]>(
    () => [
      {
        header: 'Component',
        accessorKey: 'name',
        cell: ({ row }) => (
          <span className="flex items-center gap-2">
            <ComponentType
              type={(row.original.type ?? '') as TComponentType}
              displayVariant="icon-only"
              variant="subtext"
            />
            {row.original.id === currentId ? (
              <Text variant="body" weight="strong">
                {row.original.name}
              </Text>
            ) : (
              <Link
                href={`${basePath}/${row.original.id}`}
                onClick={onNavigate}
                variant="inline"
              >
                {row.original.name}
              </Link>
            )}
          </span>
        ),
      },
      {
        header: 'Relationship',
        accessorKey: 'role',
        cell: ({ row }) => (
          <Badge size="sm" theme={ROLE_THEME[row.original.role]}>
            {ROLE_LABEL[row.original.role]}
          </Badge>
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
    ],
    [basePath, currentId, onNavigate]
  )

  return (
    <Table<TDependencyRow>
      columns={columns}
      data={rows}
      enableSearch={false}
      enableSorting={false}
      emptyStateProps={{
        emptyTitle: 'No dependencies yet',
        emptyMessage:
          'Dependencies will appear here once this component depends on or is required by other components.',
      }}
    />
  )
}
