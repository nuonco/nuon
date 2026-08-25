import { useMemo, memo } from 'react'
import { useSearchParams } from 'react-router'
import { type Node, type NodeProps } from '@xyflow/react'

import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { matchesSelector } from '@/components/match/matches'
import { cn } from '@/utils/classnames'
import type { TAppBranchConfig, TInstall } from '@/types'

import { groupAccent, type GraphAccent } from '../graph/accents'
import { GraphCanvas } from '../graph/GraphCanvas'
import { GroupNodeCard, NODE_WIDTH, NODE_WIDTH_COMPACT } from '../graph/GroupNodeCard'
import { layoutSequential, sequentialEdges } from '../graph/layout'
import { DeploymentPlanGroupPanel } from './DeploymentPlanGroupPanel'

const MAX_VISIBLE_INSTALLS = 3

export interface PlanGroupInstall {
  id: string
  name: string
  labels?: Record<string, string>
}

interface GroupNodeData {
  groupName: string
  accent: GraphAccent
  installs: PlanGroupInstall[]
  labelEntries: [string, string][]
  maxParallel: number
  useForPreviews: boolean
  compact: boolean
  orgId: string
  panelKey: string
  [key: string]: unknown
}

const GroupNode = memo(({ data }: NodeProps<Node<GroupNodeData>>) => {
  const { accent, installs, compact, orgId, panelKey } = data
  const [, setSearchParams] = useSearchParams()
  const maxVisible = compact ? 3 : MAX_VISIBLE_INSTALLS
  const visible = installs.slice(0, maxVisible)
  const hidden = installs.length - visible.length

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
      compact={compact}
      headerRight={
        <span className="flex shrink-0 items-center gap-1">
          {data.useForPreviews && (
            <span
              className={cn(
                'rounded px-1.5 py-0.5',
                compact ? 'text-[9px]' : 'text-[10px]',
                accent.pill
              )}
            >
              preview
            </span>
          )}
          {compact ? (
            <span className="text-[9px] opacity-70">{installs.length}</span>
          ) : data.maxParallel > 1 ? (
            <span className={cn('rounded px-1.5 py-0.5 text-[10px]', accent.pill)}>
              {data.maxParallel}x parallel
            </span>
          ) : null}
        </span>
      }
    >
      {!compact && data.labelEntries.length > 0 && (
        <div className="flex flex-wrap gap-1 pb-1">
          {data.labelEntries.map(([k, v]) => (
            <LabelBadge key={k} labelKey={k} labelValue={v} size="xs" />
          ))}
        </div>
      )}

      {installs.length === 0 ? (
        <span className="text-[11px] text-cool-grey-500 dark:text-cool-grey-500">
          No matching installs
        </span>
      ) : (
        <>
          {visible.map((inst) => (
            <div key={inst.id} className="flex items-center gap-1.5 min-w-0">
              <Icon
                variant="CubeIcon"
                size={compact ? 10 : 12}
                className="shrink-0 text-cool-grey-400 dark:text-cool-grey-500"
              />
              <Link
                href={`/${orgId}/installs/${inst.id}`}
                className="nodrag w-auto min-w-0 flex-1 truncate"
                title={inst.name}
              >
                {inst.name}
              </Link>
            </div>
          ))}
          {hidden > 0 &&
            (compact ? (
              <span className="text-[9px] text-cool-grey-500">+{hidden} more</span>
            ) : (
              <button
                type="button"
                onClick={openDetails}
                className="nodrag self-start text-[11px] text-cool-grey-500 hover:text-cool-grey-700 dark:hover:text-cool-grey-300"
              >
                +{hidden} installs
              </button>
            ))}
        </>
      )}

      {!compact && installs.length > 0 && (
        <DeploymentPlanGroupPanel
          panelKey={panelKey}
          groupName={data.groupName}
          installs={installs}
          orgId={orgId}
          maxParallel={data.maxParallel}
          useForPreviews={data.useForPreviews}
          labelEntries={data.labelEntries}
        />
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

export const DeploymentPlanGraph = ({ config, installsById, orgId, compact = false }: IDeploymentPlanGraph) => {
  const groups = config.install_groups ?? []

  const { nodes, edges, height } = useMemo(() => {
    if (groups.length === 0) return { nodes: [], edges: [], height: 0 }

    const built: Node<GroupNodeData>[] = groups.map((group, idx) => {
      const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})
      const installs: PlanGroupInstall[] =
        labelEntries.length > 0
          ? Object.values(installsById)
              .filter((i) => matchesSelector(i.labels, group.label_selector))
              .map((i) => ({ id: i.id, name: i.name ?? i.id, labels: i.labels }))
          : (group.install_ids ?? []).map((id) => ({
              id,
              name: installsById[id]?.name ?? id,
              labels: installsById[id]?.labels,
            }))

      const groupId = group.id || `group-${idx}`

      return {
        id: groupId,
        type: 'groupNode' as const,
        position: { x: 0, y: 0 },
        data: {
          groupName: group.name || `Group ${idx + 1}`,
          accent: groupAccent(idx),
          installs,
          labelEntries,
          maxParallel: group.max_parallel ?? 1,
          useForPreviews: group.use_for_previews ?? false,
          compact,
          orgId,
          panelKey: `install-group-plan:${groupId}`,
        },
      }
    })

    return layoutSequential(built, sequentialEdges(built.map((n) => n.id)), {
      nodeWidth: compact ? NODE_WIDTH_COMPACT : NODE_WIDTH,
      minHeight: compact ? 50 : 110,
      baseHeight: compact ? 36 : 64,
      rowHeight: compact ? 14 : 24,
      rowCount: (n) => {
        const d = n.data as GroupNodeData
        const labelRow = d.labelEntries.length > 0 ? 1 : 0
        if (compact) {
          return Math.min(d.installs.length, 4) + labelRow
        }
        const installRows =
          Math.min(d.installs.length, MAX_VISIBLE_INSTALLS) +
          (d.installs.length > MAX_VISIBLE_INSTALLS ? 1 : 0)
        const viewDetailsRow = d.installs.length > 0 ? 1 : 0
        return installRows + viewDetailsRow + labelRow
      },
      ranksep: compact ? 40 : 80,
      nodesep: compact ? 20 : 40,
    })
  }, [groups, installsById, orgId, compact])

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

  const canvasHeight = compact ? height : Math.max(height, 280)

  return (
    <GraphCanvas
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      height={canvasHeight}
      compact={compact}
      maxZoom={compact ? undefined : 1}
      fitPadding={compact ? undefined : 0.12}
    />
  )
}
