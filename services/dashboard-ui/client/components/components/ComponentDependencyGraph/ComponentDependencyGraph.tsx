import { memo, useCallback, useMemo } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  Handle,
  Position,
  MarkerType,
  type Node,
  type Edge,
  type NodeProps,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'
import { useNavigate } from 'react-router'

import { ComponentType } from '@/components/components/ComponentType'
import { useSystemTheme } from '@/hooks/use-system-theme'
import type { TComponentType } from '@/types'

const NODE_WIDTH = 200
const NODE_HEIGHT = 44

export interface GraphNode {
  id: string
  name: string
  type?: TComponentType
  role: 'current' | 'dependency' | 'dependent'
}

export interface GraphEdge {
  sourceId: string
  targetId: string
}

export interface IComponentDependencyGraph {
  nodes: GraphNode[]
  edges: GraphEdge[]
  basePath: string
  currentId: string
  onNavigate?: () => void
}

type NodeRole = 'current' | 'dependency' | 'dependent'

const ROLE_STYLES = {
  dark: {
    current: { bg: '#1e3a5f', border: '#3b82f6' },
    dependency: { bg: '#1c1c2e', border: '#4b5563' },
    dependent: { bg: '#2e1a2e', border: '#c084fc' },
  },
  light: {
    current: { bg: '#dbeafe', border: '#3b82f6' },
    dependency: { bg: '#f3f4f6', border: '#9ca3af' },
    dependent: { bg: '#f3e8ff', border: '#c084fc' },
  },
}

const EDGE_COLORS = {
  dependency: '#6b7280',
  dependent: '#c084fc',
}

const DependencyNode = memo(({ data }: NodeProps) => {
  const theme = useSystemTheme()
  const role = data.role as NodeRole
  const colors = ROLE_STYLES[theme][role]
  const isLink = role !== 'current'

  return (
    <>
      <Handle type="target" position={Position.Top} style={{ visibility: 'hidden' }} />
      <div
        className="flex items-center gap-2 px-3 py-2"
        style={{
          background: colors.bg,
          border: `2px solid ${colors.border}`,
          borderRadius: '6px',
          fontFamily: 'var(--font-hack)',
          fontSize: '12px',
          fontWeight: role === 'current' ? 600 : 500,
          minWidth: '150px',
          whiteSpace: 'nowrap',
          color: theme === 'dark' ? '#FAFAFA' : '#1f2937',
          cursor: isLink ? 'pointer' : 'default',
        }}
      >
        <ComponentType
          type={data.componentType as TComponentType}
          displayVariant="icon-only"
          variant="subtext"
        />
        <span style={isLink ? { textDecoration: 'underline', textDecorationColor: 'rgba(255,255,255,0.3)', textUnderlineOffset: '2px' } : undefined}>
          {data.label as string}
        </span>
      </div>
      <Handle type="source" position={Position.Bottom} style={{ visibility: 'hidden' }} />
    </>
  )
})
DependencyNode.displayName = 'DependencyNode'

const nodeTypes = { dependency: DependencyNode }

function layoutGraph(
  graphNodes: GraphNode[],
  graphEdges: GraphEdge[],
  currentId: string,
  basePath: string,
) {
  const nodes: Node[] = graphNodes.map((n) => ({
    id: n.id,
    type: 'dependency',
    data: {
      label: n.name,
      componentType: n.type || '',
      role: n.role,
      href: n.id !== currentId ? `${basePath}/${n.id}` : undefined,
    },
    position: { x: 0, y: 0 },
  }))

  const nodeIds = new Set(graphNodes.map((n) => n.id))
  const roleById = new Map(graphNodes.map((n) => [n.id, n.role]))

  const edges: Edge[] = graphEdges
    .filter((e) => nodeIds.has(e.sourceId) && nodeIds.has(e.targetId))
    .map((e) => {
      const targetRole = roleById.get(e.targetId)
      const color = targetRole === 'dependent' ? EDGE_COLORS.dependent : EDGE_COLORS.dependency
      return {
        id: `${e.sourceId}->${e.targetId}`,
        source: e.sourceId,
        target: e.targetId,
        type: 'smoothstep',
        style: { stroke: color, strokeWidth: 1.5 },
        markerEnd: { type: MarkerType.ArrowClosed, color },
      }
    })

  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 40, ranksep: 60 })

  nodes.forEach((node) => {
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
  })
  edges.forEach((edge) => {
    g.setEdge(edge.source, edge.target)
  })

  dagre.layout(g)

  const layoutedNodes = nodes.map((node) => {
    const pos = g.node(node.id)
    return {
      ...node,
      position: {
        x: pos.x - NODE_WIDTH / 2,
        y: pos.y - NODE_HEIGHT / 2,
      },
    }
  })

  return { nodes: layoutedNodes, edges }
}

export const ComponentDependencyGraph = ({
  nodes: graphNodes,
  edges: graphEdges,
  basePath,
  currentId,
  onNavigate,
}: IComponentDependencyGraph) => {
  const theme = useSystemTheme()
  const navigate = useNavigate()

  const { nodes, edges } = useMemo(
    () => layoutGraph(graphNodes, graphEdges, currentId, basePath),
    [graphNodes, graphEdges, currentId, basePath],
  )

  const memoizedNodeTypes = useMemo(() => nodeTypes, [])

  const onNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      const href = node.data?.href as string | undefined
      if (href) {
        navigate(href)
        onNavigate?.()
      }
    },
    [navigate, onNavigate],
  )

  return (
    <div style={{ width: '100%', height: '100%' }} className="border rounded-lg overflow-hidden">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoizedNodeTypes}
        onNodeClick={onNodeClick}
        fitView
        fitViewOptions={{ padding: 0.3 }}
        minZoom={0.5}
        maxZoom={1.5}
        proOptions={{ hideAttribution: true }}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={false}
        style={{ borderRadius: '8px' }}
      >
        <Controls
          position="top-right"
          orientation="horizontal"
          showInteractive={false}
          style={{ color: '#141217' }}
        />
        <Background
          bgColor={theme === 'dark' ? '#1D1B20' : '#FAFAFA'}
          color={theme === 'dark' ? '#333' : '#ddd'}
          gap={16}
        />
      </ReactFlow>
    </div>
  )
}
