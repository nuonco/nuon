import { useMemo, memo } from 'react'
import { type Node, type NodeProps } from '@xyflow/react'

import { LabelBadge } from '@/components/common/LabelBadge'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { cn } from '@/utils/classnames'
import type { TInstall, TInstallGroupRun } from '@/types'

import { statusAccent, type GraphAccent } from '../graph/accents'
import { GraphCanvas } from '../graph/GraphCanvas'
import { GroupNodeCard, NODE_WIDTH } from '../graph/GroupNodeCard'
import { layoutSequential, sequentialEdges } from '../graph/layout'

interface GroupRunNodeInstall {
  id: string
  name: string
  status: string
  workflowId: string
  runbooks: { name: string; status: string }[]
}

interface GroupRunNodeData {
  groupName: string
  accent: GraphAccent
  installs: GroupRunNodeInstall[]
  labelEntries: [string, string][]
  completedInstalls: number
  totalInstalls: number
  orgId: string
  [key: string]: unknown
}

const GroupRunNode = memo(({ data }: NodeProps<Node<GroupRunNodeData>>) => {
  const { accent, installs, labelEntries, orgId } = data

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
      {labelEntries.length > 0 && (
        <div className="flex flex-wrap gap-1 pb-1">
          {labelEntries.map(([k, v]) => (
            <LabelBadge key={k} labelKey={k} labelValue={v} size="xs" />
          ))}
        </div>
      )}

      {installs.length === 0 ? (
        <span className="text-[11px] text-cool-grey-500 dark:text-cool-grey-500">No installs</span>
      ) : (
        installs.map((inst) => (
          <div key={inst.id} className="flex min-w-0 flex-col gap-0.5">
            <div className="flex items-center gap-1.5 min-w-0">
              <span
                className={cn('h-1.5 w-1.5 shrink-0 rounded-full', statusAccent(inst.status).dot)}
              />
              <Link
                href={
                  inst.workflowId
                    ? `/${orgId}/installs/${inst.id}/workflows/${inst.workflowId}`
                    : `/${orgId}/installs/${inst.id}`
                }
                className="nodrag w-auto min-w-0 flex-1 truncate text-xs"
                title={inst.name}
              >
                {inst.name}
              </Link>
            </div>
            {inst.runbooks.map((rb) => (
              <div key={rb.name} className="flex items-center gap-1.5 pl-3">
                <span
                  className={cn('h-1 w-1 shrink-0 rounded-full', statusAccent(rb.status).dot)}
                />
                <span className="truncate text-[11px] text-cool-grey-500 dark:text-cool-grey-500">
                  {rb.name}
                </span>
              </div>
            ))}
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
  installsById?: Record<string, TInstall>
  orgId: string
}

export const RunDeploymentGraph = ({
  installGroupRuns,
  installsById = {},
  orgId,
}: IRunDeploymentGraph) => {
  const { nodes, edges, height } = useMemo(() => {
    if (installGroupRuns.length === 0) return { nodes: [], edges: [], height: 0 }

    const built: Node<GroupRunNodeData>[] = installGroupRuns.map((groupRun) => {
      const installs = (groupRun.installs ?? []).map((inst) => {
        const id = inst.install_id ?? ''
        return {
          id,
          name: installsById[id]?.name ?? id,
          status: inst.status ?? 'unknown',
          workflowId: inst.workflow_id ?? '',
          runbooks: (inst.runbooks ?? []).map((rb) => ({
            name: rb.runbook_name ?? rb.runbook_id ?? '',
            status: rb.status ?? 'unknown',
          })),
        }
      })

      return {
        id: groupRun.id ?? `group-${groupRun.install_group_id}`,
        type: 'groupRunNode' as const,
        position: { x: 0, y: 0 },
        data: {
          groupName: groupRun.install_group_name || 'install group',
          accent: statusAccent(groupRun.status?.status),
          installs,
          labelEntries: Object.entries(groupRun.install_group?.label_selector?.match_labels ?? {}),
          completedInstalls: groupRun.completed_installs ?? 0,
          totalInstalls: groupRun.total_installs ?? installs.length,
          orgId,
        },
      }
    })

    return layoutSequential(built, sequentialEdges(built.map((n) => n.id)), {
      nodeWidth: NODE_WIDTH,
      minHeight: 100,
      baseHeight: 60,
      rowHeight: 24,
      rowCount: (n) => {
        const d = n.data as GroupRunNodeData
        const installRows = d.installs.reduce((rows, inst) => rows + 1 + inst.runbooks.length, 0)
        return installRows + (d.labelEntries.length > 0 ? 1 : 0)
      },
    })
  }, [installGroupRuns, installsById, orgId])

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
