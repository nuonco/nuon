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

interface ComponentInfo {
  id: string
  name: string
  type?: TComponentType
}

export interface IComponentDependencyGraph {
  current: ComponentInfo
  dependencies: ComponentInfo[]
  dependents: ComponentInfo[]
  basePath: string
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

function buildGraph(
  current: ComponentInfo,
  dependencies: ComponentInfo[],
  dependents: ComponentInfo[],
  basePath: string,
) {
  const nodes: Node[] = []
  const edges: Edge[] = []

  const makeNode = (info: ComponentInfo, role: NodeRole): Node => ({
    id: info.id,
    type: 'dependency',
    data: {
      label: info.name,
      componentType: info.type || '',
      role,
      href: role !== 'current' ? `${basePath}/${info.id}` : undefined,
    },
    position: { x: 0, y: 0 },
  })

  dependencies.forEach((dep) => nodes.push(makeNode(dep, 'dependency')))
  nodes.push(makeNode(current, 'current'))
  dependents.forEach((dep) => nodes.push(makeNode(dep, 'dependent')))

  dependencies.forEach((dep) => {
    edges.push({
      id: `${dep.id}->${current.id}`,
      source: dep.id,
      target: current.id,
      type: 'smoothstep',
      style: { stroke: EDGE_COLORS.dependency, strokeWidth: 1.5 },
      markerEnd: { type: MarkerType.ArrowClosed, color: EDGE_COLORS.dependency },
    })
  })

  dependents.forEach((dep) => {
    edges.push({
      id: `${current.id}->${dep.id}`,
      source: current.id,
      target: dep.id,
      type: 'smoothstep',
      style: { stroke: EDGE_COLORS.dependent, strokeWidth: 1.5 },
      markerEnd: { type: MarkerType.ArrowClosed, color: EDGE_COLORS.dependent },
    })
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
  current,
  dependencies,
  dependents,
  basePath,
  onNavigate,
}: IComponentDependencyGraph) => {
  const theme = useSystemTheme()
  const navigate = useNavigate()

  const { nodes, edges } = useMemo(
    () => buildGraph(current, dependencies, dependents, basePath),
    [current, dependencies, dependents, basePath],
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
