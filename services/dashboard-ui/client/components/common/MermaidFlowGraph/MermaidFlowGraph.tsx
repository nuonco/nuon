import { memo, useMemo } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  Handle,
  Position,
  MarkerType,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
  type NodeProps,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'
import { useSystemTheme } from '@/hooks/use-system-theme'
import { parseMermaidFlowchart, type ParsedNode, type ParsedEdge, type ParsedSubgraph } from './parse-mermaid'

const MIN_NODE_WIDTH = 120
const MAX_NODE_WIDTH = 260
const CHAR_WIDTH = 7.5
const NODE_PADDING_X = 24
const NODE_HEIGHT = 44
const SUBGRAPH_PADDING = 30

const THEME = {
  dark: {
    nodeBg: '#141217',
    nodeText: '#FFFFFF',
    canvasBg: '#1D1B20',
    dotColor: '#FFFFFF',
    edgeColor: '#6B7280',
    edgeLabelText: '#D1D5DB',
    edgeLabelBg: '#1D1B20',
    groupBg: 'rgba(255,255,255,0.05)',
    groupBorder: 'rgba(255,255,255,0.15)',
    groupText: 'rgba(255,255,255,0.6)',
  },
  light: {
    nodeBg: '#FFFFFF',
    nodeText: '#141217',
    canvasBg: '#EAEDF0',
    dotColor: '#141217',
    edgeColor: '#9CA3AF',
    edgeLabelText: '#4B5563',
    edgeLabelBg: '#EAEDF0',
    groupBg: 'rgba(0,0,0,0.03)',
    groupBorder: 'rgba(0,0,0,0.12)',
    groupText: 'rgba(0,0,0,0.5)',
  },
} as const

function measureNodeWidth(label: string): number {
  const longestLine = label.split('\n').reduce((max, line) => Math.max(max, line.length), 0)
  const measured = longestLine * CHAR_WIDTH + NODE_PADDING_X
  return Math.min(MAX_NODE_WIDTH, Math.max(MIN_NODE_WIDTH, measured))
}

type ShapeConfig = {
  borderRadius?: number | string
  clipPath?: string
  borderWidth?: number
  sizeMultiplier?: number
  svgShape?: boolean
}

const shapeConfigs: Record<string, ShapeConfig> = {
  rect: { borderRadius: 4 },
  rounded: { borderRadius: 999 },
  diamond: { clipPath: 'polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%)', sizeMultiplier: 1.6 },
  cylinder: { svgShape: true, sizeMultiplier: 1.4 },
  circle: { borderRadius: '50%', sizeMultiplier: 1.2 },
  subroutine: { borderRadius: 4, borderWidth: 3 },
  asymmetric: { clipPath: 'polygon(0% 0%, 90% 0%, 100% 50%, 90% 100%, 0% 100%)' },
  parallelogram: { clipPath: 'polygon(10% 0%, 100% 0%, 90% 100%, 0% 100%)' },
}

function getNodeDimensions(shape: string, baseWidth: number): { width: number; height: number } {
  const config = shapeConfigs[shape] || shapeConfigs.rect
  const mult = config.sizeMultiplier || 1
  if (shape === 'circle') {
    const size = Math.max(baseWidth, NODE_HEIGHT) * mult
    return { width: size, height: size }
  }
  return { width: baseWidth * mult, height: NODE_HEIGHT * mult }
}

function getNodeStyle(
  node: ParsedNode,
  width: number,
  height: number,
  theme: 'dark' | 'light',
): React.CSSProperties {
  const colors = THEME[theme]
  const bgColor = node.style?.fill || colors.nodeBg
  const borderColor = node.style?.stroke || 'var(--border-color)'
  const textColor = node.style?.color || colors.nodeText
  const config = shapeConfigs[node.nodeShape] || shapeConfigs.rect

  const noBorder = config.clipPath || config.svgShape

  return {
    backgroundColor: noBorder ? 'transparent' : bgColor,
    border: noBorder ? 'none' : `${config.borderWidth || 1}px solid ${borderColor}`,
    ...(config.svgShape ? { '--node-fill': bgColor, '--node-stroke': borderColor } as React.CSSProperties : {}),
    color: textColor,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    width,
    height,
    fontSize: 11,
    fontFamily: 'var(--font-hack)',
    padding: '4px 8px',
    textAlign: 'center',
    lineHeight: '1.3',
    ...(config.borderRadius != null ? { borderRadius: config.borderRadius } : {}),
    ...(config.clipPath ? { clipPath: config.clipPath, backgroundColor: bgColor } : {}),
  }
}

const CylinderNode = ({ label, style }: { label: string; style: React.CSSProperties }) => {
  const w = Number(style.width) || 120
  const h = Number(style.height) || 56
  const ry = 8
  const fill = (style as any)['--node-fill'] || '#141217'
  const stroke = (style as any)['--node-stroke'] || 'var(--border-color)'

  return (
    <div style={{ width: w, height: h, position: 'relative' }}>
      <svg
        width={w}
        height={h}
        viewBox={`0 0 ${w} ${h}`}
        style={{ position: 'absolute', top: 0, left: 0 }}
      >
        <path
          d={`M 0,${ry} Q 0,0 ${w / 2},0 Q ${w},0 ${w},${ry} L ${w},${h - ry} Q ${w},${h} ${w / 2},${h} Q 0,${h} 0,${h - ry} Z`}
          fill={fill}
          stroke={stroke}
          strokeWidth={1}
        />
        <ellipse
          cx={w / 2}
          cy={ry}
          rx={w / 2}
          ry={ry}
          fill={fill}
          stroke={stroke}
          strokeWidth={1}
        />
      </svg>
      <div
        style={{
          position: 'relative',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: '100%',
          height: '100%',
          paddingTop: ry,
          fontSize: style.fontSize,
          fontFamily: style.fontFamily,
          color: style.color,
          textAlign: 'center',
          lineHeight: style.lineHeight,
        }}
      >
        <span>
          {label.split('\n').map((line, i) => (
            <span key={i}>
              {i > 0 && <br />}
              {line}
            </span>
          ))}
        </span>
      </div>
    </div>
  )
}

const NodeLabel = ({ label }: { label: string }) => (
  <span>
    {label.split('\n').map((line, i) => (
      <span key={i}>
        {i > 0 && <br />}
        {line}
      </span>
    ))}
  </span>
)

const MermaidNode = memo(({ data }: NodeProps) => {
  const { label, nodeStyle, isCylinder } = data as { label: string; nodeStyle: React.CSSProperties; isCylinder?: boolean }

  return (
    <>
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <Handle type="target" position={Position.Left} style={{ opacity: 0 }} />
      {isCylinder ? (
        <CylinderNode label={label} style={nodeStyle} />
      ) : (
        <div style={nodeStyle}>
          <NodeLabel label={label} />
        </div>
      )}
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
      <Handle type="source" position={Position.Right} style={{ opacity: 0 }} />
    </>
  )
})

MermaidNode.displayName = 'MermaidNode'

const SubgraphLabel = memo(({ data }: NodeProps) => {
  const theme = useSystemTheme()
  const colors = THEME[theme]
  const { label, width, height } = data as { label: string; width: number; height: number }

  return (
    <div
      style={{
        width,
        height,
        backgroundColor: colors.groupBg,
        border: `1px dashed ${colors.groupBorder}`,
        borderRadius: 8,
        padding: '6px 10px',
        fontSize: 11,
        fontWeight: 600,
        color: colors.groupText,
        pointerEvents: 'none',
      }}
    >
      {label}
    </div>
  )
})

SubgraphLabel.displayName = 'SubgraphLabel'

const nodeTypes = { mermaid: MermaidNode, subgraphBox: SubgraphLabel }

function collectAllChildren(sg: ParsedSubgraph, subgraphs: ParsedSubgraph[]): string[] {
  const all = [...sg.children]
  for (const childId of sg.children) {
    const childSg = subgraphs.find((s) => s.id === childId)
    if (childSg) {
      all.push(...collectAllChildren(childSg, subgraphs))
    }
  }
  return all
}

function buildLayout(
  parsedNodes: ParsedNode[],
  parsedEdges: ParsedEdge[],
  subgraphs: ParsedSubgraph[],
  direction: string,
  theme: 'dark' | 'light',
  colors: typeof THEME['dark'],
): { nodes: Node[]; edges: Edge[] } {
  if (parsedNodes.length === 0) {
    return { nodes: [], edges: [] }
  }

  const nodeDims = new Map<string, { width: number; height: number }>()
  for (const node of parsedNodes) {
    const baseWidth = measureNodeWidth(node.label)
    nodeDims.set(node.id, getNodeDimensions(node.nodeShape, baseWidth))
  }

  const nodeMap = new Map(parsedNodes.map((n) => [n.id, n]))
  const allNodeIds = new Set(parsedNodes.map((n) => n.id))

  // Build parent→children mapping (direct children only, excluding nested subgraph children)
  const directChildNodes = new Map<string, string[]>()
  const childSubgraphs = new Map<string, string[]>()
  const sgIds = new Set(subgraphs.map((s) => s.id))

  for (const sg of subgraphs) {
    const nodes: string[] = []
    const subs: string[] = []
    for (const c of sg.children) {
      if (sgIds.has(c)) subs.push(c)
      else if (allNodeIds.has(c)) nodes.push(c)
    }
    directChildNodes.set(sg.id, nodes)
    childSubgraphs.set(sg.id, subs)
  }

  // Find top-level nodes (not in any subgraph)
  const nodesInAnySg = new Set<string>()
  for (const sg of subgraphs) {
    for (const c of sg.children) {
      if (allNodeIds.has(c)) nodesInAnySg.add(c)
    }
  }
  // Find top-level subgraphs (not children of any other subgraph)
  const sgsInAnySg = new Set<string>()
  for (const sg of subgraphs) {
    for (const c of sg.children) {
      if (sgIds.has(c)) sgsInAnySg.add(c)
    }
  }

  // Layout a set of items (nodes + subgraph boxes) with edges between them
  type LayoutItem = { id: string; width: number; height: number }
  type LayoutResult = Map<string, { x: number; y: number }>

  function layoutItems(items: LayoutItem[], edges: Array<{ source: string; target: string }>): LayoutResult {
    const g = new dagre.graphlib.Graph()
    g.setDefaultEdgeLabel(() => ({}))
    g.setGraph({ rankdir: direction, nodesep: 50, ranksep: 80 })
    for (const item of items) {
      g.setNode(item.id, { width: item.width, height: item.height })
    }
    for (const edge of edges) {
      if (g.hasNode(edge.source) && g.hasNode(edge.target)) {
        g.setEdge(edge.source, edge.target)
      }
    }
    dagre.layout(g)
    const result: LayoutResult = new Map()
    for (const item of items) {
      const n = g.node(item.id)
      if (n) result.set(item.id, { x: n.x - item.width / 2, y: n.y - item.height / 2 })
    }
    return result
  }

  // Recursively layout subgraphs from leaves up
  // Returns: { positions (absolute), boxSize }
  type SgLayout = { positions: Map<string, { x: number; y: number }>; width: number; height: number }
  const sgLayouts = new Map<string, SgLayout>()

  function layoutSubgraph(sgId: string): SgLayout {
    if (sgLayouts.has(sgId)) return sgLayouts.get(sgId)!

    const childNodeIds = directChildNodes.get(sgId) || []
    const childSgIds = childSubgraphs.get(sgId) || []

    // Layout child subgraphs first
    for (const csId of childSgIds) layoutSubgraph(csId)

    // Build items: real nodes + subgraph boxes
    const items: LayoutItem[] = []
    for (const nid of childNodeIds) {
      const dims = nodeDims.get(nid)!
      items.push({ id: nid, ...dims })
    }
    for (const csId of childSgIds) {
      const csLayout = sgLayouts.get(csId)!
      items.push({ id: csId, width: csLayout.width, height: csLayout.height })
    }

    // Collect all descendant node IDs for edge filtering
    const allDescendantNodes = new Set(childNodeIds)
    for (const csId of childSgIds) {
      const desc = collectAllChildren(subgraphs.find((s) => s.id === csId)!, subgraphs)
      desc.forEach((id) => { if (allNodeIds.has(id)) allDescendantNodes.add(id) })
    }

    // Edges internal to this subgraph (both endpoints are descendants)
    // Map descendant nodes to their direct parent item in this layout
    const nodeToItem = new Map<string, string>()
    for (const nid of childNodeIds) nodeToItem.set(nid, nid)
    for (const csId of childSgIds) {
      const desc = collectAllChildren(subgraphs.find((s) => s.id === csId)!, subgraphs)
      desc.forEach((id) => { if (allNodeIds.has(id)) nodeToItem.set(id, csId) })
    }

    const internalEdges: Array<{ source: string; target: string }> = []
    for (const e of parsedEdges) {
      const si = nodeToItem.get(e.source)
      const ti = nodeToItem.get(e.target)
      if (si && ti && si !== ti && allDescendantNodes.has(e.source) && allDescendantNodes.has(e.target)) {
        internalEdges.push({ source: si, target: ti })
      }
    }

    if (items.length === 0) {
      const result = { positions: new Map(), width: 0, height: 0 }
      sgLayouts.set(sgId, result)
      return result
    }

    const localPositions = layoutItems(items, internalEdges)

    // Compute bounding box
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity
    for (const item of items) {
      const pos = localPositions.get(item.id)
      if (!pos) continue
      minX = Math.min(minX, pos.x)
      minY = Math.min(minY, pos.y)
      maxX = Math.max(maxX, pos.x + item.width)
      maxY = Math.max(maxY, pos.y + item.height)
    }

    // Normalize to 0,0 origin with padding
    const pad = SUBGRAPH_PADDING
    const labelH = 22
    const offsetX = -minX + pad
    const offsetY = -minY + pad + labelH

    const positions = new Map<string, { x: number; y: number }>()

    // Place direct child nodes
    for (const nid of childNodeIds) {
      const lp = localPositions.get(nid)
      if (lp) positions.set(nid, { x: lp.x + offsetX, y: lp.y + offsetY })
    }

    // Place child subgraph contents (offset their positions)
    for (const csId of childSgIds) {
      const csLayout = sgLayouts.get(csId)!
      const lp = localPositions.get(csId)
      if (!lp) continue
      const ox = lp.x + offsetX
      const oy = lp.y + offsetY
      // Store subgraph box position
      positions.set(csId, { x: ox, y: oy })
      // Offset all child node positions
      for (const [nid, npos] of csLayout.positions) {
        positions.set(nid, { x: npos.x + ox, y: npos.y + oy })
      }
    }

    const totalW = (maxX - minX) + pad * 2
    const totalH = (maxY - minY) + pad * 2 + labelH

    const result = { positions, width: totalW, height: totalH }
    sgLayouts.set(sgId, result)
    return result
  }

  // Layout all top-level subgraphs
  const topSgs = subgraphs.filter((sg) => !sgsInAnySg.has(sg.id))
  for (const sg of topSgs) layoutSubgraph(sg.id)

  // Top-level layout: orphan nodes + subgraph boxes
  const topItems: LayoutItem[] = []
  const topOrphanNodes = parsedNodes.filter((n) => !nodesInAnySg.has(n.id))
  for (const n of topOrphanNodes) {
    topItems.push({ id: n.id, ...nodeDims.get(n.id)! })
  }
  for (const sg of topSgs) {
    const layout = sgLayouts.get(sg.id)!
    topItems.push({ id: sg.id, width: layout.width, height: layout.height })
  }

  // Top-level edges: map node→top-level item
  const nodeToTopItem = new Map<string, string>()
  for (const n of topOrphanNodes) nodeToTopItem.set(n.id, n.id)
  for (const sg of topSgs) {
    const allDesc = collectAllChildren(sg, subgraphs)
    allDesc.forEach((id) => { if (allNodeIds.has(id)) nodeToTopItem.set(id, sg.id) })
  }

  const topEdges: Array<{ source: string; target: string }> = []
  for (const e of parsedEdges) {
    const si = nodeToTopItem.get(e.source)
    const ti = nodeToTopItem.get(e.target)
    if (si && ti && si !== ti) topEdges.push({ source: si, target: ti })
  }

  const topPositions = layoutItems(topItems, topEdges)

  // Assemble final absolute positions
  const positions = new Map<string, { x: number; y: number }>()

  for (const n of topOrphanNodes) {
    const p = topPositions.get(n.id)
    if (p) positions.set(n.id, p)
  }

  for (const sg of topSgs) {
    const sgPos = topPositions.get(sg.id)
    if (!sgPos) continue
    positions.set(sg.id, sgPos)
    const layout = sgLayouts.get(sg.id)!
    for (const [nid, npos] of layout.positions) {
      positions.set(nid, { x: npos.x + sgPos.x, y: npos.y + sgPos.y })
    }
  }

  // Build ReactFlow nodes
  const nodes: Node[] = []

  // Subgraph boxes (render inner-most first for z-ordering)
  const allSgsReversed = [...subgraphs].reverse()
  for (const sg of allSgsReversed) {
    const sgPos = positions.get(sg.id)
    const layout = sgLayouts.get(sg.id)
    if (!sgPos || !layout) continue

    nodes.push({
      id: sg.id,
      type: 'subgraphBox',
      data: { label: sg.label, width: layout.width, height: layout.height },
      position: sgPos,
      zIndex: -1,
      selectable: false,
      draggable: false,
      connectable: false,
    })
  }

  for (const node of parsedNodes) {
    const pos = positions.get(node.id)
    if (!pos) continue
    const dims = nodeDims.get(node.id)!

    nodes.push({
      id: node.id,
      type: 'mermaid',
      data: { label: node.label, nodeStyle: getNodeStyle(node, dims.width, dims.height, theme), isCylinder: node.nodeShape === 'cylinder' },
      position: pos,
      style: { background: 'transparent', border: 'none', padding: 0, width: dims.width, height: dims.height },
      zIndex: 1,
    })
  }

  const validNodeIds = new Set(parsedNodes.map((n) => n.id))
  const edges: Edge[] = parsedEdges
    .filter((e) => validNodeIds.has(e.source) && validNodeIds.has(e.target))
    .map((e, i) => ({
      id: `e-${e.source}-${e.target}-${i}`,
      source: e.source,
      target: e.target,
      type: 'smoothstep',
      label: e.label,
      labelStyle: { fontSize: 9, fontFamily: 'var(--font-hack)', fill: colors.edgeLabelText },
      labelBgStyle: { fill: colors.edgeLabelBg, fillOpacity: 0.85 },
      labelBgPadding: [4, 2] as [number, number],
      labelBgBorderRadius: 3,
      animated: false,
      style: {
        stroke: 'var(--mermaid-edge-color)',
        strokeWidth: e.type === 'thick' ? 3 : 2,
        ...(e.type === 'dashed' ? { strokeDasharray: '5,5' } : {}),
      },
      ...(e.hasArrow
        ? { markerEnd: { type: MarkerType.ArrowClosed, color: 'var(--mermaid-edge-color)' } }
        : {}),
    }))

  return { nodes, edges }
}

export const MermaidFlowGraph = ({ code }: { code: string }) => {
  const theme = useSystemTheme()
  const colors = THEME[theme]

  const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(() => {
    const parsed = parseMermaidFlowchart(code)
    return buildLayout(parsed.nodes, parsed.edges, parsed.subgraphs, parsed.direction, theme, colors)
  }, [code, theme])

  const [nodes, , onNodesChange] = useNodesState(layoutedNodes as Node[])
  const [edges, , onEdgesChange] = useEdgesState(layoutedEdges)

  const memoizedNodeTypes = useMemo(() => nodeTypes, [])

  return (
    <div
      className="w-full h-[44rem] my-4 border rounded-lg border-color"
      style={{
        '--mermaid-edge-color': colors.edgeColor,
      } as React.CSSProperties}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoizedNodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
        fitViewOptions={{ padding: 0.15 }}
        minZoom={0.2}
        maxZoom={1.5}
        proOptions={{ hideAttribution: true }}
        nodesDraggable={false}
        nodesConnectable={false}
        style={{ borderRadius: '8px' }}
      >
        <Controls
          position="top-right"
          orientation="horizontal"
          style={{ color: '#141217' }}
        />
        <Background bgColor={colors.canvasBg} color={colors.dotColor} gap={16} />
      </ReactFlow>
    </div>
  )
}
