import { useEffect, useMemo, memo } from 'react'
import {
  ReactFlow,
  Node,
  Edge,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  MarkerType,
  Handle,
  Position,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'

const NODE_WIDTH = 200
const NODE_HEIGHT = 40

function getLayoutedElements(nodes: Node[], edges: Edge[], direction = 'LR') {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: direction, nodesep: 60, ranksep: 100 })

  nodes.forEach((node) => g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT }))
  edges.forEach((edge) => g.setEdge(edge.source, edge.target))

  dagre.layout(g)

  return {
    nodes: nodes.map((node) => {
      const pos = g.node(node.id)
      return { ...node, position: { x: pos.x - NODE_WIDTH / 2, y: pos.y - NODE_HEIGHT / 2 } }
    }),
    edges,
  }
}

const CustomNode = memo(({ data, id }: any) => {
  const bg = data.color === 'blue' ? '#1e50c0' : '#662F9D'
  return (
    <>
      <Handle type="target" position={Position.Left} style={{ background: '#555' }} />
      <div
        style={{
          background: bg,
          color: '#FAFAFA',
          borderRadius: '6px',
          fontSize: '11px',
          fontWeight: 500,
          fontFamily: 'ui-monospace, monospace',
          padding: '8px 12px',
          minWidth: '120px',
          textAlign: 'center',
          whiteSpace: 'nowrap',
          boxShadow: '0 1px 3px rgba(0,0,0,0.3)',
        }}
      >
        {String(data.label || id).replace(/\\n/g, '\n')}
      </div>
      <Handle type="source" position={Position.Right} style={{ background: '#555' }} />
    </>
  )
})
CustomNode.displayName = 'CustomNode'

const nodeTypes = { custom: CustomNode }

function parseDot(dot: string): { nodes: Node[]; edges: Edge[] } {
  const nodesMap = new Map<string, Node>()
  const edges: Edge[] = []
  const allIds = new Set<string>()

  // Parse node declarations: "id" [label="...", ...];
  // Also match: "id" [attrs]; and id [attrs];
  const nodeRe = /^\s*"?([^"\s\[]+)"?\s*\[\s*([^\]]+?)\s*\];?\s*$/gm
  let m: RegExpExecArray | null
  while ((m = nodeRe.exec(dot)) !== null) {
    const [, rawId, attrs] = m
    const id = rawId.trim()
    // Skip graph-level attributes like "node [...]" or "graph [...]"
    if (id === 'node' || id === 'graph' || id === 'edge') continue

    allIds.add(id)
    const a: Record<string, string> = {}
    const aRe = /(\w+)\s*=\s*"([^"]*)"/g
    let am: RegExpExecArray | null
    while ((am = aRe.exec(attrs)) !== null) a[am[1]] = am[2]

    nodesMap.set(id, {
      id,
      type: 'custom',
      data: { label: a.label || a.name || id, color: a.color || 'purple' },
      position: { x: 0, y: 0 },
    })
  }

  // Parse edges: "source" -> "target" or "source" -> "target" [attrs];
  const edgeRe = /^\s*"?([^"\s]+)"?\s*->\s*"?([^"\s\[;]+)"?\s*(?:\[\s*([^\]]*)\s*\])?\s*;?\s*$/gm
  while ((m = edgeRe.exec(dot)) !== null) {
    const [, rawSrc, rawTgt, attrs] = m
    const src = rawSrc.trim()
    const tgt = rawTgt.trim()
    if (src === 'node' || src === 'graph' || tgt === 'node' || tgt === 'graph') continue

    allIds.add(src)
    allIds.add(tgt)

    const a: Record<string, string> = {}
    if (attrs) {
      const aRe = /(\w+)\s*=\s*"([^"]*)"/g
      let am: RegExpExecArray | null
      while ((am = aRe.exec(attrs)) !== null) a[am[1]] = am[2]
    }

    const color = a.color === 'red' ? '#991B1B' : '#8040BF'
    edges.push({
      id: `${src}-${tgt}`,
      source: src,
      target: tgt,
      type: 'smoothstep',
      style: { stroke: color, strokeWidth: 2 },
      markerEnd: { type: MarkerType.ArrowClosed, color },
    })
  }

  // Add implicit nodes (referenced in edges but not declared)
  allIds.forEach((id) => {
    if (!nodesMap.has(id)) {
      nodesMap.set(id, {
        id,
        type: 'custom',
        data: { label: id, color: 'purple' },
        position: { x: 0, y: 0 },
      })
    }
  })

  return getLayoutedElements(Array.from(nodesMap.values()), edges)
}

interface IDotGraph {
  dot: string
  height?: string
}

export const DotGraph = ({ dot, height = '24rem' }: IDotGraph) => {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

  useEffect(() => {
    if (dot) {
      const { nodes: n, edges: e } = parseDot(dot)
      setNodes(n)
      setEdges(e)
    }
  }, [dot, setNodes, setEdges])

  const memoTypes = useMemo(() => nodeTypes, [])

  if (!dot) return null
  if (nodes.length === 0) {
    return (
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm text-gray-500 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-400">
        <p>No graph nodes found. Raw DOT:</p>
        <pre className="mt-2 text-xs font-mono overflow-x-auto max-h-32">{dot}</pre>
      </div>
    )
  }

  return (
    <div className="w-full border border-gray-200 rounded-lg overflow-hidden dark:border-gray-800" style={{ height }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
        fitViewOptions={{ padding: 0.3 }}
        minZoom={0.1}
        maxZoom={2}
        proOptions={{ hideAttribution: true }}
      >
        <Controls position="top-right" orientation="horizontal" />
        <Background bgColor="#1B242C" color="#444" gap={16} />
      </ReactFlow>
    </div>
  )
}
