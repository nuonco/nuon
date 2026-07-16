import { useMemo, memo } from 'react'
import { type Node, type NodeProps } from '@xyflow/react'

import { Link } from '@/components/common/Link'
import { EmptyState } from '@/components/common/EmptyState'
import { cn } from '@/utils/classnames'
import type { TInstallGroupRun } from '@/types'

import { statusAccent, type GraphAccent } from '../graph/accents'
import { GraphCanvas } from '../graph/GraphCanvas'
import { GroupNodeCard, NODE_WIDTH } from '../graph/GroupNodeCard'
import { layoutSequential, sequentialEdges } from '../graph/layout'

interface GroupRunNodeData {
  groupName: string
  accent: GraphAccent
  installs: { id: string; status: string }[]
  orgId: string
  completedInstalls: number
  totalInstalls: number
  [key: string]: unknown
}

const GroupRunNode = memo(({ data }: NodeProps<Node<GroupRunNodeData>>) => {
  const { accent, installs, orgId } = data

  return (
    <GroupNodeCard
      accent={accent}
      title={data.groupName}
      headerRight={
        <span className={cn('rounded px-1.5 py-0.5 text-[10px]', accent.pill)}>
          {data.completedInstalls}/{data.totalInstalls}
        </span>
      }
    >
      {installs.length === 0 ? (
        <span className="text-[11px] text-cool-grey-500 dark:text-cool-grey-500">No installs</span>
      ) : (
        installs.map((inst) => (
          <div key={inst.id} className="flex items-center gap-1.5">
            <span
              className={cn('h-1.5 w-1.5 shrink-0 rounded-full', statusAccent(inst.status).dot)}
            />
            <Link
              href={`/${orgId}/installs/${inst.id}`}
              className="truncate text-xs text-cool-grey-700 dark:text-cool-grey-200"
            >
              {inst.id}
            </Link>
          </div>
        ))
      )}
    </GroupNodeCard>
  )
})

GroupRunNode.displayName = 'GroupRunNode'

const nodeTypes = { groupRunNode: GroupRunNode }

export interface IRunDeploymentGraph {
  installGroupRuns: TInstallGroupRun[]
  orgId: string
}

export const RunDeploymentGraph = ({ installGroupRuns, orgId }: IRunDeploymentGraph) => {
  const { nodes, edges, height } = useMemo(() => {
    if (installGroupRuns.length === 0) return { nodes: [], edges: [], height: 0 }

    const built: Node<GroupRunNodeData>[] = installGroupRuns.map((groupRun) => {
      const installs = (groupRun.installs ?? []).map((inst) => ({
        id: inst.install_id ?? '',
        status: inst.status ?? 'unknown',
      }))

      return {
        id: groupRun.id ?? `group-${groupRun.install_group_id}`,
        type: 'groupRunNode' as const,
        position: { x: 0, y: 0 },
        data: {
          groupName: groupRun.install_group_name || 'install group',
          accent: statusAccent(groupRun.status?.status),
          installs,
          orgId,
          completedInstalls: groupRun.completed_installs ?? 0,
          totalInstalls: groupRun.total_installs ?? installs.length,
        },
      }
    })

    return layoutSequential(built, sequentialEdges(built.map((n) => n.id)), {
      nodeWidth: NODE_WIDTH,
      minHeight: 100,
      baseHeight: 60,
      rowHeight: 24,
      rowCount: (n) => (n.data as GroupRunNodeData).installs.length,
    })
  }, [installGroupRuns, orgId])

  if (installGroupRuns.length === 0) {
    return (
      <div className="rounded-lg border p-6">
        <EmptyState
          variant="diagram"
          emptyTitle="No installs updated"
          emptyMessage="No install groups were deployed during this run."
        />
      </div>
    )
  }

  return <GraphCanvas nodes={nodes} edges={edges} nodeTypes={nodeTypes} height={height} />
}
