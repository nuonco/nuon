import { useMemo, memo } from 'react'
import { useSearchParams } from 'react-router'
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
import { GroupRunDetailPanel } from './GroupRunDetailPanel'

const MAX_VISIBLE_INSTALLS = 3
const MAX_VISIBLE_RUNBOOKS = 2

export interface GroupRunInstall {
  id: string
  name: string
  status: string
  workflowId: string
  runbooks: { name: string; status: string }[]
}

function distinctRunbookNames(installs: GroupRunInstall[]): string[] {
  const names = new Set<string>()
  for (const inst of installs) {
    for (const rb of inst.runbooks) {
      if (rb.name) names.add(rb.name)
    }
  }
  return Array.from(names)
}

interface GroupRunNodeData {
  groupName: string
  status?: string
  accent: GraphAccent
  installs: GroupRunInstall[]
  runbookNames: string[]
  labelEntries: [string, string][]
  completedInstalls: number
  totalInstalls: number
  orgId: string
  panelKey: string
  [key: string]: unknown
}

const GroupRunNode = memo(({ data }: NodeProps<Node<GroupRunNodeData>>) => {
  const { accent, installs, runbookNames, labelEntries, orgId, panelKey } = data
  const [, setSearchParams] = useSearchParams()
  const visible = installs.slice(0, MAX_VISIBLE_INSTALLS)
  const hidden = installs.length - visible.length
  const visibleRunbooks = runbookNames.slice(0, MAX_VISIBLE_RUNBOOKS)
  const hiddenRunbooks = runbookNames.length - visibleRunbooks.length

  const openDetails = () => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.set('panel', panelKey)
        return next
      },
      { replace: false }
    )
  }

  return (
    <GroupNodeCard
      accent={accent}
      title={data.groupName}
      headerRight={
        <span className={cn('shrink-0 rounded px-1.5 py-0.5 text-[10px]', accent.pill)}>
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
        <>
          {visible.map((inst) => (
            <div key={inst.id} className="flex items-center gap-1.5 min-w-0">
              <span
                className={cn('h-1.5 w-1.5 shrink-0 rounded-full', statusAccent(inst.status).dot)}
              />
              <Link
                href={
                  inst.workflowId
                    ? `/${orgId}/installs/${inst.id}/workflows/${inst.workflowId}`
                    : `/${orgId}/installs/${inst.id}`
                }
                className="nodrag w-auto min-w-0 flex-1 truncate"
                title={inst.name}
              >
                {inst.name}
              </Link>
            </div>
          ))}
          {hidden > 0 && (
            <button
              type="button"
              onClick={openDetails}
              className="nodrag self-start text-[11px] text-cool-grey-500 hover:text-cool-grey-700 dark:hover:text-cool-grey-300"
            >
              +{hidden} installs
            </button>
          )}
        </>
      )}

      {runbookNames.length > 0 && (
        <div className="mt-1 flex flex-col gap-1 border-t pt-2">
          <span className="text-[10px] font-strong uppercase tracking-wide text-cool-grey-500 dark:text-cool-grey-500">
            Runbooks
          </span>
          {visibleRunbooks.map((name) => (
            <span key={name} className="min-w-0 truncate text-xs" title={name}>
              {name}
            </span>
          ))}
          {hiddenRunbooks > 0 && (
            <button
              type="button"
              onClick={openDetails}
              className="nodrag self-start text-[11px] text-cool-grey-500 hover:text-cool-grey-700 dark:hover:text-cool-grey-300"
            >
              +{hiddenRunbooks} runbooks
            </button>
          )}
        </div>
      )}

      {installs.length > 0 && (
        <GroupRunDetailPanel
          panelKey={panelKey}
          groupName={data.groupName}
          status={data.status}
          completedInstalls={data.completedInstalls}
          totalInstalls={data.totalInstalls}
          installs={installs}
          orgId={orgId}
        />
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
      const installs: GroupRunInstall[] = (groupRun.installs ?? []).map((inst) => {
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

      const runbookNames = distinctRunbookNames(installs)
      const labelEntries = Object.entries(
        groupRun.install_group?.label_selector?.match_labels ?? {}
      )

      const groupRunId = groupRun.id ?? `group-${groupRun.install_group_id}`

      return {
        id: groupRunId,
        type: 'groupRunNode' as const,
        position: { x: 0, y: 0 },
        data: {
          groupName: groupRun.install_group_name || 'install group',
          status: groupRun.status?.status,
          accent: statusAccent(groupRun.status?.status),
          installs,
          runbookNames,
          labelEntries,
          completedInstalls: groupRun.completed_installs ?? 0,
          totalInstalls: groupRun.total_installs ?? installs.length,
          orgId,
          panelKey: `install-group-run:${groupRunId}`,
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
        const installRows =
          Math.min(d.installs.length, MAX_VISIBLE_INSTALLS) +
          (d.installs.length > MAX_VISIBLE_INSTALLS ? 1 : 0)
        const runbookRows =
          d.runbookNames.length > 0
            ? Math.min(d.runbookNames.length, MAX_VISIBLE_RUNBOOKS) +
              (d.runbookNames.length > MAX_VISIBLE_RUNBOOKS ? 1 : 0) +
              1
            : 0
        const viewDetailsRow = d.installs.length > 0 ? 1 : 0
        return (
          installRows +
          runbookRows +
          viewDetailsRow +
          (d.labelEntries.length > 0 ? 1 : 0)
        )
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
