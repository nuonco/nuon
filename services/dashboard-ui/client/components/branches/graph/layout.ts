import dagre from '@dagrejs/dagre'
import { MarkerType, type Node, type Edge } from '@xyflow/react'

export interface LayoutOptions {
  nodeWidth: number
  minHeight: number
  baseHeight: number
  rowHeight: number
  rowCount: (node: Node) => number
  ranksep?: number
  nodesep?: number
}

const MARGIN_X = 32
const MARGIN_Y = 24
const HEIGHT_PADDING = 12

const nodeHeight = (node: Node, opts: LayoutOptions) =>
  Math.max(
    opts.minHeight,
    opts.baseHeight + opts.rowCount(node) * opts.rowHeight
  )

export function layoutSequential(
  nodes: Node[],
  edges: Edge[],
  opts: LayoutOptions
): { nodes: Node[]; edges: Edge[]; height: number } {
  const graph = new dagre.graphlib.Graph()
  graph.setDefaultEdgeLabel(() => ({}))
  graph.setGraph({
    rankdir: 'LR',
    ranksep: opts.ranksep ?? 80,
    nodesep: opts.nodesep ?? 40,
  })

  for (const node of nodes) {
    graph.setNode(node.id, {
      width: opts.nodeWidth,
      height: nodeHeight(node, opts),
    })
  }
  for (const edge of edges) {
    graph.setEdge(edge.source, edge.target)
  }

  dagre.layout(graph)

  const positioned = nodes.map((node) => {
    const pos = graph.node(node.id)
    return {
      ...node,
      position: {
        x: pos.x - opts.nodeWidth / 2,
        y: pos.y - nodeHeight(node, opts) / 2,
      },
    }
  })

  const minX = Math.min(...positioned.map((n) => n.position.x))
  const minY = Math.min(...positioned.map((n) => n.position.y))
  for (const node of positioned) {
    node.position.x -= minX - MARGIN_X
    node.position.y -= minY - MARGIN_Y
  }

  const height =
    Math.max(...positioned.map((n) => n.position.y + nodeHeight(n, opts))) +
    HEIGHT_PADDING

  return { nodes: positioned, edges, height }
}

export function sequentialEdges(ids: string[]): Edge[] {
  const edges: Edge[] = []
  for (let i = 0; i < ids.length - 1; i++) {
    edges.push({
      id: `${ids[i]}-${ids[i + 1]}`,
      source: ids[i],
      target: ids[i + 1],
      type: 'smoothstep',
      animated: true,
      style: {
        stroke: 'var(--foreground)',
        strokeOpacity: 0.3,
        strokeWidth: 2,
      },
      markerEnd: { type: MarkerType.ArrowClosed, color: 'var(--foreground)' },
    })
  }
  return edges
}
