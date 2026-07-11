import { useMemo, useEffect, memo } from 'react'
import {
  ReactFlow,
  type Node,
  type Edge,
  Controls,
  type NodeProps,
  Handle,
  Position,
  MarkerType,
  useNodesState,
  useEdgesState,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'

import { EmptyState } from '@/components/common/EmptyState'
import { matchesSelector } from '@/components/match/matches'
import type { TAppBranchConfig, TInstall } from '@/types'

const GROUP_COLORS = [
  { bg: '#1e3a5f', border: '#3b82f6', text: '#93c5fd' },
  { bg: '#1a3d2e', border: '#22c55e', text: '#86efac' },
  { bg: '#3d2e1a', border: '#f59e0b', text: '#fcd34d' },
  { bg: '#2e1a3d', border: '#a855f7', text: '#d8b4fe' },
  { bg: '#3d1a2e', border: '#ec4899', text: '#f9a8d4' },
  { bg: '#1a2e3d', border: '#06b6d4', text: '#67e8f9' },
]

const NODE_WIDTH = 280
const NODE_WIDTH_COMPACT = 160
const NODE_MIN_HEIGHT = 100
const NODE_MIN_HEIGHT_COMPACT = 50

interface GroupNodeData {
  groupName: string
  colorIndex: number
  installs: { id: string; name: string; install?: TInstall }[]
  labelEntries: [string, string][]
  maxParallel: number
  orgId: string
  compact?: boolean
  [key: string]: unknown
}

const GroupNode = memo(({ data }: NodeProps<Node<GroupNodeData>>) => {
  const color = GROUP_COLORS[data.colorIndex % GROUP_COLORS.length]
  const installs = data.installs ?? []
  const compact = data.compact

  if (compact) {
    return (
      <>
        <Handle type="target" position={Position.Left} style={{ opacity: 0 }} />
        <div
          className="rounded overflow-hidden"
          style={{
            background: color.bg,
            border: `1px solid ${color.border}`,
            minWidth: NODE_WIDTH_COMPACT,
            fontFamily: 'var(--font-hack)',
            fontSize: '10px',
          }}
        >
          <div
            className="px-2 py-1 flex items-center justify-between gap-1"
            style={{ borderBottom: `1px solid ${color.border}40` }}
          >
            <span className="truncate" style={{ color: color.text, fontWeight: 600 }}>{data.groupName}</span>
            <span className="text-[9px] shrink-0" style={{ color: `${color.text}99` }}>
              {installs.length}
            </span>
          </div>
          <div className="px-2 py-1 flex flex-col gap-0.5">
            {installs.slice(0, 3).map((inst) => (
              <div key={inst.id} className="flex items-center gap-1">
                <span className="w-1 h-1 rounded-full shrink-0" style={{ background: color.border }} />
                <span className="truncate text-[9px]" style={{ color: '#d1d5db' }}>{inst.name}</span>
              </div>
            ))}
            {installs.length > 3 && (
              <span className="text-[9px]" style={{ color: `${color.text}70` }}>+{installs.length - 3} more</span>
            )}
            {installs.length === 0 && (
              <span className="text-[9px]" style={{ color: `${color.text}60` }}>empty</span>
            )}
          </div>
        </div>
        <Handle type="source" position={Position.Right} style={{ opacity: 0 }} />
      </>
    )
  }

  return (
    <>
      <Handle type="target" position={Position.Left} style={{ opacity: 0 }} />
      <div
        className="rounded-lg overflow-hidden"
        style={{
          background: color.bg,
          border: `2px solid ${color.border}`,
          minWidth: NODE_WIDTH,
          fontFamily: 'var(--font-hack)',
          fontSize: '12px',
        }}
      >
        <div
          className="px-3 py-2 flex items-center justify-between gap-2"
          style={{ borderBottom: `1px solid ${color.border}` }}
        >
          <span style={{ color: color.text, fontWeight: 600 }}>{data.groupName}</span>
          {data.maxParallel > 1 && (
            <span
              className="text-[10px] px-1.5 py-0.5 rounded"
              style={{ background: color.border, color: '#fff' }}
            >
              {data.maxParallel}x parallel
            </span>
          )}
        </div>

        {(data.labelEntries?.length ?? 0) > 0 && (
          <div className="px-3 py-1.5 flex flex-wrap gap-1" style={{ borderBottom: `1px solid ${color.border}20` }}>
            {data.labelEntries.map(([k, v]) => (
              <span
                key={k}
                className="text-[10px] px-1.5 py-0.5 rounded"
                style={{ background: `${color.border}30`, color: color.text }}
              >
                {k}={v}
              </span>
            ))}
          </div>
        )}

        <div className="px-3 py-2 flex flex-col gap-1">
          {installs.length === 0 ? (
            <span className="text-[11px]" style={{ color: `${color.text}80` }}>
              No matching installs
            </span>
          ) : (
            installs.map((inst) => (
              <div key={inst.id} className="flex items-center gap-1.5">
                <span
                  className="w-1.5 h-1.5 rounded-full shrink-0"
                  style={{ background: color.border }}
                />
                <span className="truncate" style={{ color: '#e5e7eb' }}>
                  {inst.name}
                </span>
              </div>
            ))
          )}
        </div>
      </div>
      <Handle type="source" position={Position.Right} style={{ opacity: 0 }} />
    </>
  )
})

GroupNode.displayName = 'GroupNode'

const nodeTypes = { groupNode: GroupNode }

function getLayoutedElements(nodes: Node[], edges: Edge[], opts?: { compact?: boolean }) {
  const w = opts?.compact ? NODE_WIDTH_COMPACT : NODE_WIDTH
  const minH = opts?.compact ? NODE_MIN_HEIGHT_COMPACT : NODE_MIN_HEIGHT
  const installRowH = opts?.compact ? 14 : 24
  const baseH = opts?.compact ? 36 : 60
  const ranksep = opts?.compact ? 40 : 80
  const nodesep = opts?.compact ? 20 : 40

  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))
  dagreGraph.setGraph({ rankdir: 'LR', ranksep, nodesep })

  nodes.forEach((node) => {
    const installCount = Math.min((node.data as GroupNodeData).installs?.length ?? 0, opts?.compact ? 4 : 999)
    const height = Math.max(minH, baseH + installCount * installRowH)
    dagreGraph.setNode(node.id, { width: w, height })
  })

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target)
  })

  dagre.layout(dagreGraph)

  const positioned = nodes.map((node) => {
    const pos = dagreGraph.node(node.id)
    const installCount = Math.min((node.data as GroupNodeData).installs?.length ?? 0, opts?.compact ? 4 : 999)
    const height = Math.max(minH, baseH + installCount * installRowH)
    return {
      ...node,
      position: { x: pos.x - w / 2, y: pos.y - height / 2 },
    }
  })

  const minX = Math.min(...positioned.map((n) => n.position.x))
  for (const node of positioned) {
    node.position.x -= minX - 32
  }

  const minY = Math.min(...positioned.map((n) => n.position.y))
  for (const node of positioned) {
    node.position.y -= minY - 25
  }

  return { nodes: positioned, edges }
}

interface IDeploymentPlanGraph {
  config: TAppBranchConfig
  installsById: Record<string, TInstall>
  orgId: string
  compact?: boolean
}

export const DeploymentPlanGraph = ({
  config,
  installsById,
  orgId,
  compact = false,
}: IDeploymentPlanGraph) => {
  const groups = config.install_groups ?? []
  const nodeWidth = compact ? NODE_WIDTH_COMPACT : NODE_WIDTH
  const nodeMinHeight = compact ? NODE_MIN_HEIGHT_COMPACT : NODE_MIN_HEIGHT

  const { nodes: initialNodes, edges: initialEdges, graphHeight } = useMemo(() => {
    if (groups.length === 0) return { nodes: [] as Node[], edges: [] as Edge[] }

    const nodes: Node<GroupNodeData>[] = groups.map((group, idx) => {
      const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})
      const isLabels = labelEntries.length > 0

      const installs = isLabels
        ? Object.values(installsById)
            .filter((i) => matchesSelector(i.labels, group.label_selector))
            .map((i) => ({ id: i.id, name: i.name ?? i.id, install: i }))
        : (group.install_ids ?? []).map((id) => ({
            id,
            name: installsById[id]?.name ?? id,
            install: installsById[id],
          }))

      return {
        id: group.id || `group-${idx}`,
        type: 'groupNode' as const,
        position: { x: 0, y: 0 },
        data: {
          groupName: group.name || `Group ${idx + 1}`,
          colorIndex: idx,
          installs,
          labelEntries,
          maxParallel: group.max_parallel ?? 1,
          orgId,
          compact,
        },
      }
    })

    const edges: Edge[] = []
    for (let i = 0; i < nodes.length - 1; i++) {
      const color = GROUP_COLORS[i % GROUP_COLORS.length].border
      edges.push({
        id: `${nodes[i].id}-${nodes[i + 1].id}`,
        source: nodes[i].id,
        target: nodes[i + 1].id,
        type: 'smoothstep',
        animated: true,
        style: { stroke: color, strokeWidth: 2 },
        markerEnd: { type: MarkerType.ArrowClosed, color },
      })
    }

    const laid = getLayoutedElements(nodes, edges, { compact })

    const installRowH = compact ? 14 : 24
    const baseH = compact ? 36 : 60
    const minH = compact ? NODE_MIN_HEIGHT_COMPACT : NODE_MIN_HEIGHT

    let maxBottom = 0
    for (const node of laid.nodes) {
      const ic = Math.min((node.data as GroupNodeData).installs?.length ?? 0, compact ? 4 : 999)
      const nh = Math.max(minH, baseH + ic * installRowH)
      maxBottom = Math.max(maxBottom, node.position.y + nh)
    }

    return { ...laid, graphHeight: maxBottom + 10 }
  }, [groups, installsById, orgId, compact])

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)

  useEffect(() => {
    setNodes(initialNodes)
    setEdges(initialEdges)
  }, [initialNodes, initialEdges, setNodes, setEdges])

  const memoizedNodeTypes = useMemo(() => nodeTypes, [])

  if (groups.length === 0) {
    return (
      <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg p-6">
        <EmptyState
          variant="diagram"
          emptyTitle="No install groups configured"
          emptyMessage={`Use "Deployment plan" above to add deployment groups.`}
        />
      </div>
    )
  }

  return (
    <div
      className={compact ? 'w-full border rounded bg-dark-grey-900 dark:bg-dark-grey-900' : 'w-full border rounded-lg bg-dark-grey-900 dark:bg-dark-grey-900'}
      style={{ height: `${graphHeight}px` }}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoizedNodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        defaultViewport={{ x: 0, y: 0, zoom: 1 }}
        minZoom={compact ? 0.6 : 0.5}
        maxZoom={compact ? 1 : 1.5}
        proOptions={{ hideAttribution: true }}
        style={{ borderRadius: compact ? '4px' : '8px' }}
      >
        {!compact && (
          <Controls
            position="top-right"
            orientation="horizontal"
            style={{ color: '#121212' }}
          />
        )}
      </ReactFlow>
    </div>
  )
}
