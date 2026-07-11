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

import { Link } from '@/components/common/Link'
import { EmptyState } from '@/components/common/EmptyState'
import type { TInstallGroupRun } from '@/types'

const STATUS_COLORS: Record<string, { bg: string; border: string; text: string }> = {
  success: { bg: '#1a3d2e', border: '#22c55e', text: '#86efac' },
  error: { bg: '#3d1a1a', border: '#ef4444', text: '#fca5a5' },
  'in-progress': { bg: '#1e3a5f', border: '#3b82f6', text: '#93c5fd' },
  pending: { bg: '#2a2a2a', border: '#6b7280', text: '#9ca3af' },
}

const FALLBACK_COLOR = { bg: '#2a2a2a', border: '#6b7280', text: '#9ca3af' }

const INSTALL_STATUS_DOT: Record<string, string> = {
  success: '#22c55e',
  error: '#ef4444',
  'in-progress': '#3b82f6',
  pending: '#6b7280',
}

const NODE_WIDTH = 280
const NODE_MIN_HEIGHT = 100

interface GroupRunNodeData {
  groupName: string
  status: string
  installs: { id: string; status: string; workflowId?: string }[]
  orgId: string
  completedInstalls: number
  totalInstalls: number
  failedInstalls: number
  [key: string]: unknown
}

const GroupRunNode = memo(({ data }: NodeProps<Node<GroupRunNodeData>>) => {
  const color = STATUS_COLORS[data.status] || FALLBACK_COLOR
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
          <span
            className="text-[10px] px-1.5 py-0.5 rounded"
            style={{ background: color.border, color: '#fff' }}
          >
            {data.completedInstalls}/{data.totalInstalls}
          </span>
        </div>

        <div className="px-3 py-2 flex flex-col gap-1">
          {installs.length === 0 ? (
            <span className="text-[11px]" style={{ color: `${color.text}80` }}>
              No installs
            </span>
          ) : (
            installs.map((inst) => (
              <div key={inst.id} className="flex items-center justify-between gap-2">
                <div className="flex items-center gap-1.5 min-w-0">
                  <span
                    className="w-1.5 h-1.5 rounded-full shrink-0"
                    style={{ background: INSTALL_STATUS_DOT[inst.status] || '#6b7280' }}
                  />
                  <Link
                    href={`/${data.orgId}/installs/${inst.id}`}
                    className="truncate text-[12px]"
                    style={{ color: '#e5e7eb' }}
                  >
                    {inst.id}
                  </Link>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
      <Handle type="source" position={Position.Right} style={{ opacity: 0 }} />
    </>
  )
})

GroupRunNode.displayName = 'GroupRunNode'

const nodeTypes = { groupRunNode: GroupRunNode }

function getLayoutedElements(nodes: Node[], edges: Edge[]) {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))
  dagreGraph.setGraph({ rankdir: 'LR', ranksep: 80, nodesep: 40 })

  nodes.forEach((node) => {
    const installCount = ((node.data as GroupRunNodeData).installs?.length ?? 0)
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
      const installCount = ((node.data as GroupRunNodeData).installs?.length ?? 0)
      const height = Math.max(NODE_MIN_HEIGHT, 60 + installCount * 24)
      return {
        ...node,
        position: { x: pos.x - NODE_WIDTH / 2, y: pos.y - height / 2 },
      }
    }),
    edges,
  }
}

export interface IRunDeploymentGraph {
  installGroupRuns: TInstallGroupRun[]
  orgId: string
}

export const RunDeploymentGraph = ({
  installGroupRuns,
  orgId,
}: IRunDeploymentGraph) => {
  const { nodes: initialNodes, edges: initialEdges } = useMemo(() => {
    if (installGroupRuns.length === 0) return { nodes: [] as Node[], edges: [] as Edge[] }

    const nodes: Node<GroupRunNodeData>[] = installGroupRuns.map((groupRun) => {
      const installs = (groupRun.installs ?? []).map((inst) => ({
        id: inst.install_id ?? '',
        status: inst.status ?? 'unknown',
        workflowId: inst.workflow_id,
      }))

      const status = groupRun.status?.status ?? 'pending'

      return {
        id: groupRun.id ?? `group-${groupRun.install_group_id}`,
        type: 'groupRunNode' as const,
        position: { x: 0, y: 0 },
        data: {
          groupName: groupRun.install_group_name || 'install group',
          status,
          installs,
          orgId,
          completedInstalls: groupRun.completed_installs ?? 0,
          totalInstalls: groupRun.total_installs ?? installs.length,
          failedInstalls: groupRun.failed_installs ?? 0,
        },
      }
    })

    const edges: Edge[] = []
    for (let i = 0; i < nodes.length - 1; i++) {
      const color = (STATUS_COLORS[nodes[i].data.status] || FALLBACK_COLOR).border
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
  }, [installGroupRuns, orgId])

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)

  useEffect(() => {
    setNodes(initialNodes)
    setEdges(initialEdges)
  }, [initialNodes, initialEdges, setNodes, setEdges])

  const memoizedNodeTypes = useMemo(() => nodeTypes, [])

  if (installGroupRuns.length === 0) {
    return (
      <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg p-6">
        <EmptyState
          variant="diagram"
          emptyTitle="No installs updated"
          emptyMessage="No install groups were deployed during this run."
        />
      </div>
    )
  }

  const maxInstalls = Math.max(
    ...installGroupRuns.map((g) => g.installs?.length ?? 0),
    0
  )
  const graphHeight = Math.max(250, 100 + maxInstalls * 28 + 40)

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
