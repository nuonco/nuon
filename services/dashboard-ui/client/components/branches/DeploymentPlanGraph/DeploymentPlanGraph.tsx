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
const NODE_MIN_HEIGHT = 100

interface GroupNodeData {
  groupName: string
  colorIndex: number
  installs: { id: string; name: string; install?: TInstall }[]
  labelEntries: [string, string][]
  maxParallel: number
  orgId: string
  [key: string]: unknown
}

const GroupNode = memo(({ data }: NodeProps<Node<GroupNodeData>>) => {
  const color = GROUP_COLORS[data.colorIndex % GROUP_COLORS.length]
  const installs = data.installs ?? []

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

function getLayoutedElements(nodes: Node[], edges: Edge[]) {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))
  dagreGraph.setGraph({ rankdir: 'LR', ranksep: 80, nodesep: 40 })

  nodes.forEach((node) => {
    const installCount = ((node.data as GroupNodeData).installs?.length ?? 0)
    const height = Math.max(NODE_MIN_HEIGHT, 60 + installCount * 24)
    dagreGraph.setNode(node.id, { width: NODE_WIDTH, height })
  })

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target)
  })

  dagre.layout(dagreGraph)

  return {
    nodes: nodes.map((node) => {
      const pos = dagreGraph.node(node.id)
      const installCount = ((node.data as GroupNodeData).installs?.length ?? 0)
      const height = Math.max(NODE_MIN_HEIGHT, 60 + installCount * 24)
      return {
        ...node,
        position: { x: pos.x - NODE_WIDTH / 2, y: pos.y - height / 2 },
      }
    }),
    edges,
  }
}

interface IDeploymentPlanGraph {
  config: TAppBranchConfig
  installsById: Record<string, TInstall>
  orgId: string
}

export const DeploymentPlanGraph = ({
  config,
  installsById,
  orgId,
}: IDeploymentPlanGraph) => {
  const groups = config.install_groups ?? []

  const { nodes: initialNodes, edges: initialEdges } = useMemo(() => {
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

    return getLayoutedElements(nodes, edges)
  }, [groups, installsById, orgId])

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

  const maxInstalls = Math.max(...groups.map((g) => {
    const labelEntries = Object.entries(g.label_selector?.match_labels ?? {})
    if (labelEntries.length > 0) {
      return Object.values(installsById).filter((i) => matchesSelector(i.labels, g.label_selector)).length
    }
    return (g.install_ids ?? []).length
  }), 0)
  const graphHeight = Math.max(300, 100 + maxInstalls * 28 + 40)

  return (
    <div
      className="w-full border rounded-lg bg-dark-grey-900 dark:bg-dark-grey-900"
      style={{ height: `${graphHeight}px` }}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoizedNodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
        fitViewOptions={{ padding: 0.3 }}
        minZoom={0.5}
        maxZoom={1.5}
        proOptions={{ hideAttribution: true }}
        style={{ borderRadius: '8px' }}
      >
        <Controls
          position="top-right"
          orientation="horizontal"
          style={{ color: '#121212' }}
        />
      </ReactFlow>
    </div>
  )
}
