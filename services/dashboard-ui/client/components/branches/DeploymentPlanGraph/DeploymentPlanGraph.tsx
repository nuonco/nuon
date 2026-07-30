import { useMemo, memo } from 'react'
import { type Node, type NodeProps } from '@xyflow/react'

import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { matchesSelector } from '@/components/match/matches'
import { cn } from '@/utils/classnames'
import type { TAppBranchConfig, TInstall } from '@/types'

import { groupAccent, type GraphAccent } from '../graph/accents'
import { GraphCanvas } from '../graph/GraphCanvas'
import { GroupNodeCard, NODE_WIDTH, NODE_WIDTH_COMPACT } from '../graph/GroupNodeCard'
import { layoutSequential, sequentialEdges } from '../graph/layout'

interface GroupNodeData {
  groupName: string
  accent: GraphAccent
  installs: { id: string; name: string }[]
  labelEntries: [string, string][]
  maxParallel: number
  compact: boolean
  [key: string]: unknown
}

const GroupNode = memo(({ data }: NodeProps<Node<GroupNodeData>>) => {
  const { accent, installs, compact } = data
  const visible = compact ? installs.slice(0, 3) : installs

  return (
    <GroupNodeCard
      accent={accent}
      title={data.groupName}
      compact={compact}
      headerRight={
        compact ? (
          <span className="shrink-0 text-[9px] opacity-70">{installs.length}</span>
        ) : data.maxParallel > 1 ? (
          <span className={cn('rounded px-1.5 py-0.5 text-[10px]', accent.pill)}>
            {data.maxParallel}x parallel
          </span>
        ) : null
      }
    >
      {!compact && data.labelEntries.length > 0 && (
        <div className="flex flex-wrap gap-1 pb-1">
          {data.labelEntries.map(([k, v]) => (
            <span key={k} className={cn('rounded px-1.5 py-0.5 text-[10px]', accent.pill)}>
              {k}={v}
            </span>
          ))}
        </div>
      )}

      {installs.length === 0 ? (
        <span className="text-[11px] text-cool-grey-500 dark:text-cool-grey-500">
          No matching installs
        </span>
      ) : (
        visible.map((inst) => (
          <div key={inst.id} className="flex items-center gap-1.5">
            <Icon
              variant="CubeIcon"
              size={compact ? 10 : 12}
              className="shrink-0 text-cool-grey-400 dark:text-cool-grey-500"
            />
            <span
              className={cn(
                'truncate text-cool-grey-700 dark:text-cool-grey-200',
                compact ? 'text-[9px]' : 'text-xs'
              )}
            >
              {inst.name}
            </span>
          </div>
        ))
      )}
      {compact && installs.length > 3 && (
        <span className="text-[9px] text-cool-grey-500">+{installs.length - 3} more</span>
      )}
    </GroupNodeCard>
  )
})

GroupNode.displayName = 'GroupNode'

const nodeTypes = { groupNode: GroupNode }

interface IDeploymentPlanGraph {
  config: TAppBranchConfig
  installsById: Record<string, TInstall>
  orgId: string
  compact?: boolean
}

export const DeploymentPlanGraph = ({ config, installsById, compact = false }: IDeploymentPlanGraph) => {
  const groups = config.install_groups ?? []

  const { nodes, edges, height } = useMemo(() => {
    if (groups.length === 0) return { nodes: [], edges: [], height: 0 }

    const built: Node<GroupNodeData>[] = groups.map((group, idx) => {
      const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})
      const installs =
        labelEntries.length > 0
          ? Object.values(installsById)
              .filter((i) => matchesSelector(i.labels, group.label_selector))
              .map((i) => ({ id: i.id, name: i.name ?? i.id }))
          : (group.install_ids ?? []).map((id) => ({
              id,
              name: installsById[id]?.name ?? id,
            }))

      return {
        id: group.id || `group-${idx}`,
        type: 'groupNode' as const,
        position: { x: 0, y: 0 },
        data: {
          groupName: group.name || `Group ${idx + 1}`,
          accent: groupAccent(idx),
          installs,
          labelEntries,
          maxParallel: group.max_parallel ?? 1,
          compact,
        },
      }
    })

    return layoutSequential(built, sequentialEdges(built.map((n) => n.id)), {
      nodeWidth: compact ? NODE_WIDTH_COMPACT : NODE_WIDTH,
      minHeight: compact ? 50 : 100,
      baseHeight: compact ? 36 : 60,
      rowHeight: compact ? 14 : 24,
      rowCount: (n) => {
        const d = n.data as GroupNodeData
        const rows = compact ? Math.min(d.installs.length, 4) : d.installs.length
        return rows + (d.labelEntries.length > 0 ? 1 : 0)
      },
      ranksep: compact ? 40 : 80,
      nodesep: compact ? 20 : 40,
    })
  }, [groups, installsById, compact])

  if (groups.length === 0) {
    return (
      <div className="rounded-lg border p-6">
        <EmptyState
          variant="diagram"
          size={compact ? 'sm' : 'default'}
          emptyTitle="No install groups"
          emptyMessage="This deployment plan doesn't have any install groups yet."
        />
      </div>
    )
  }

  return <GraphCanvas nodes={nodes} edges={edges} nodeTypes={nodeTypes} height={height} compact={compact} />
}
