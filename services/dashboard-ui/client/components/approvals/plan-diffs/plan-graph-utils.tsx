import { memo } from 'react'
import {
  Handle,
  Position,
  MarkerType,
  type Node,
  type Edge,
  type NodeProps,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'

export const STRUCTURAL_COLOR = '#374151'
export const DRIFT_STRUCTURAL_COLOR = '#92400e'

export const ACTION_COLORS: Record<string, string> = {
  create: '#16a34a',
  delete: '#dc2626',
  update: '#ea580c',
  replace: '#8040BF',
  'create-replacement': '#8040BF',
  'delete-replaced': '#8040BF',
  read: '#2563eb',
  refresh: '#2563eb',
  'no-op': '#6b7280',
  same: '#6b7280',
}

export function estimateNodeWidth(label: string) {
  return Math.max(120, label.length * 8 + 24)
}

export function getLayoutedElements(nodes: Node[], edges: Edge[]) {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 40, ranksep: 60 })

  const nodeHeight = 40

  nodes.forEach((node) => {
    const w = estimateNodeWidth(node.data.label as string)
    g.setNode(node.id, { width: w, height: nodeHeight })
  })
  edges.forEach((edge) => {
    g.setEdge(edge.source, edge.target)
  })

  dagre.layout(g)

  return nodes.map((node) => {
    const w = estimateNodeWidth(node.data.label as string)
    const pos = g.node(node.id)
    return {
      ...node,
      position: { x: pos.x - w / 2, y: pos.y - nodeHeight / 2 },
    }
  })
}

export function createAddNode(nodes: Node[], nodeIds: Set<string>) {
  return (id: string, type: string, data: Record<string, unknown>) => {
    if (nodeIds.has(id)) return
    nodeIds.add(id)
    const w = estimateNodeWidth(data.label as string)
    nodes.push({
      id,
      type,
      data,
      position: { x: 0, y: 0 },
      width: w,
      style: { background: 'transparent', border: 'none', padding: 0, width: w },
    })
  }
}

export function createAddEdge(edges: Edge[]) {
  return (source: string, target: string, isDrift = false) => {
    const color = isDrift ? '#d97706' : '#6b7280'
    edges.push({
      id: `${source}->${target}`,
      source,
      target,
      type: 'smoothstep',
      style: { stroke: color, strokeWidth: 1.5, strokeDasharray: isDrift ? '6 3' : undefined },
      markerEnd: { type: MarkerType.ArrowClosed, color },
    })
  }
}

const nodeStyle = (bg: string, fontSize: string, fontWeight: number, border?: string) => ({
  background: bg,
  color: '#FAFAFA',
  borderRadius: '6px',
  fontFamily: 'var(--font-hack)',
  fontSize,
  fontWeight,
  minWidth: '120px',
  textAlign: 'center' as const,
  whiteSpace: 'nowrap' as const,
  border: border || 'none',
})

export const StructuralNode = memo(({ data }: NodeProps) => {
  const isDrift = data.isDrift as boolean
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ visibility: 'hidden' }} />
      <div
        className="flex items-center justify-center px-3 py-2"
        style={nodeStyle(
          isDrift ? DRIFT_STRUCTURAL_COLOR : STRUCTURAL_COLOR,
          '12px',
          600,
          isDrift ? '2px dashed #d97706' : undefined,
        )}
        title={data.label as string}
      >
        {data.label as string}
      </div>
      <Handle type="source" position={Position.Bottom} style={{ visibility: 'hidden' }} />
    </>
  )
})
StructuralNode.displayName = 'StructuralNode'

export const ActionNode = memo(({ data }: NodeProps) => {
  const action = data.action as string
  const isDrift = data.isDrift as boolean
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ visibility: 'hidden' }} />
      <div
        className="flex items-center justify-center px-3 py-2"
        style={nodeStyle(
          ACTION_COLORS[action] || '#6b7280',
          '11px',
          500,
          isDrift ? '2px dashed #d97706' : undefined,
        )}
        title={data.label as string}
      >
        {data.label as string}
      </div>
    </>
  )
})
ActionNode.displayName = 'ActionNode'

export const graphNodeTypes = {
  structural: StructuralNode,
  action: ActionNode,
}
