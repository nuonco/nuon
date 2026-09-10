import ELK, { type ElkNode } from 'elkjs/lib/elk.bundled.js'

export type TGraphDirection = 'right' | 'down'

export interface IGraphLayoutNode {
  id: string
  parentId?: string
  width: number
  height: number
}

export interface IGraphLayoutEdge {
  id: string
  source: string
  target: string
}

export interface IGraphLayoutInput<
  TNode extends IGraphLayoutNode = IGraphLayoutNode,
> {
  nodes: TNode[]
  edges: IGraphLayoutEdge[]
  direction?: TGraphDirection
}

export interface IGraphPoint {
  x: number
  y: number
}

export type TPositionedGraphNode<TNode extends IGraphLayoutNode> = TNode & {
  position: IGraphPoint
}

export type TRoutedGraphEdge<
  TEdge extends IGraphLayoutEdge = IGraphLayoutEdge,
> = TEdge & {
  points: IGraphPoint[]
}

export interface IGraphLayoutResult<
  TNode extends IGraphLayoutNode,
  TEdge extends IGraphLayoutEdge,
> {
  nodes: TPositionedGraphNode<TNode>[]
  edges: TRoutedGraphEdge<TEdge>[]
}

const CONTAINER_PADDING = 24
const CONTAINER_HEADER = 44
const NODE_SPACING = 40
const LAYER_SPACING = 80
const EDGE_NODE_SPACING = 24
const EDGE_SPACING = 16
const CORNER_RADIUS = 8

const DIRECTION_OPTIONS: Record<TGraphDirection, string> = {
  right: 'RIGHT',
  down: 'DOWN',
}

const layoutOptions = (direction: TGraphDirection) => ({
  'elk.algorithm': 'layered',
  'elk.direction': DIRECTION_OPTIONS[direction],
  'elk.hierarchyHandling': 'INCLUDE_CHILDREN',
  'elk.edgeRouting': 'ORTHOGONAL',
  'elk.padding': `[top=${CONTAINER_HEADER},left=${CONTAINER_PADDING},bottom=${CONTAINER_PADDING},right=${CONTAINER_PADDING}]`,
  'elk.spacing.nodeNode': `${NODE_SPACING}`,
  'elk.spacing.edgeNode': `${EDGE_NODE_SPACING}`,
  'elk.spacing.edgeEdge': `${EDGE_SPACING}`,
  'elk.layered.spacing.nodeNodeBetweenLayers': `${LAYER_SPACING}`,
  'elk.layered.spacing.edgeNodeBetweenLayers': `${EDGE_NODE_SPACING}`,
  'elk.layered.spacing.edgeEdgeBetweenLayers': `${EDGE_SPACING}`,
})

const validParentId = (
  node: IGraphLayoutNode,
  nodesById: Map<string, IGraphLayoutNode>
) => {
  if (!node.parentId || !nodesById.has(node.parentId)) return undefined

  const visited = new Set([node.id])
  let ancestorId: string | undefined = node.parentId

  while (ancestorId) {
    if (visited.has(ancestorId)) return undefined
    visited.add(ancestorId)
    ancestorId = nodesById.get(ancestorId)?.parentId
  }

  return node.parentId
}

const toElkChildren = (
  parentId: string | undefined,
  childrenByParent: Map<string | undefined, IGraphLayoutNode[]>,
  options: Record<string, string>
): ElkNode[] =>
  (childrenByParent.get(parentId) ?? []).map((node) => {
    const children = toElkChildren(node.id, childrenByParent, options)

    return {
      id: node.id,
      width: node.width,
      height: node.height,
      children,
      ...(children.length > 0 ? { layoutOptions: options } : {}),
    }
  })

export const graphLayoutSignature = ({
  nodes,
  edges,
}: Pick<IGraphLayoutInput, 'nodes' | 'edges'>) =>
  JSON.stringify({
    nodes: nodes
      .map(({ id, parentId }) => [id, parentId ?? ''])
      .sort(([leftId], [rightId]) => leftId.localeCompare(rightId)),
    edges: edges
      .map(({ source, target }) => [source, target])
      .sort(([leftSource, leftTarget], [rightSource, rightTarget]) =>
        `${leftSource}:${leftTarget}`.localeCompare(
          `${rightSource}:${rightTarget}`
        )
      ),
  })

export const orthogonalEdgePath = (
  points: IGraphPoint[],
  radius = CORNER_RADIUS
) => {
  if (points.length === 0) return ''
  const [first, ...rest] = points
  if (rest.length === 0) return `M ${first!.x} ${first!.y}`

  const distance = (from: IGraphPoint, to: IGraphPoint) =>
    Math.hypot(to.x - from.x, to.y - from.y)

  const towards = (from: IGraphPoint, to: IGraphPoint, by: number) => {
    const length = distance(from, to)
    if (length === 0) return from
    const ratio = Math.min(by, length / 2) / length
    return {
      x: from.x + (to.x - from.x) * ratio,
      y: from.y + (to.y - from.y) * ratio,
    }
  }

  let path = `M ${first!.x} ${first!.y}`

  for (let index = 1; index < points.length - 1; index += 1) {
    const previous = points[index - 1]!
    const corner = points[index]!
    const next = points[index + 1]!
    const entry = towards(corner, previous, radius)
    const exit = towards(corner, next, radius)

    path += ` L ${entry.x} ${entry.y} Q ${corner.x} ${corner.y} ${exit.x} ${exit.y}`
  }

  const last = points.at(-1)!
  return `${path} L ${last.x} ${last.y}`
}

export const layoutGraph = async <
  TNode extends IGraphLayoutNode,
  TEdge extends IGraphLayoutEdge,
>({
  nodes,
  edges,
  direction = 'right',
}: IGraphLayoutInput<TNode> & { edges: TEdge[] }): Promise<
  IGraphLayoutResult<TNode, TEdge>
> => {
  const nodesById = new Map(nodes.map((node) => [node.id, node]))
  const parentIds = new Map(
    nodes.map((node) => [node.id, validParentId(node, nodesById)])
  )
  const childrenByParent = new Map<string | undefined, IGraphLayoutNode[]>()

  for (const node of nodes) {
    const parentId = parentIds.get(node.id)
    childrenByParent.set(parentId, [
      ...(childrenByParent.get(parentId) ?? []),
      node,
    ])
  }

  const options = layoutOptions(direction)
  const elk = new ELK()
  const result = await elk.layout({
    id: 'root',
    layoutOptions: options,
    children: toElkChildren(undefined, childrenByParent, options),
    edges: edges.map(({ id, source, target }) => ({
      id,
      sources: [source],
      targets: [target],
    })),
  })

  type TElkEdge = NonNullable<ElkNode['edges']>[number]

  const positioned: TPositionedGraphNode<TNode>[] = []
  const origins = new Map<string, IGraphPoint>([['root', { x: 0, y: 0 }]])
  const elkEdges = new Map<string, { edge: TElkEdge; owner: string }>()

  const collect = (elkNodes: ElkNode[] | undefined, origin: IGraphPoint) => {
    for (const elkNode of elkNodes ?? []) {
      const node = nodesById.get(elkNode.id)
      const position = { x: elkNode.x ?? 0, y: elkNode.y ?? 0 }
      const absolute = {
        x: origin.x + position.x,
        y: origin.y + position.y,
      }
      origins.set(elkNode.id, absolute)

      if (node) {
        positioned.push({
          ...node,
          parentId: parentIds.get(node.id),
          width: elkNode.width ?? node.width,
          height: elkNode.height ?? node.height,
          position,
        })
      }

      for (const elkEdge of elkNode.edges ?? []) {
        elkEdges.set(elkEdge.id, { edge: elkEdge, owner: elkNode.id })
      }

      collect(elkNode.children, absolute)
    }
  }

  for (const elkEdge of result.edges ?? []) {
    elkEdges.set(elkEdge.id, { edge: elkEdge, owner: 'root' })
  }
  collect(result.children, { x: 0, y: 0 })

  const routed = edges.map((edge): TRoutedGraphEdge<TEdge> => {
    const found = elkEdges.get(edge.id)
    const section = found?.edge.sections?.at(0)
    const containerId = found?.edge.container ?? found?.owner ?? 'root'
    const origin = origins.get(containerId) ?? { x: 0, y: 0 }
    const points = section
      ? [
          section.startPoint,
          ...(section.bendPoints ?? []),
          section.endPoint,
        ].map((point) => ({
          x: origin.x + point.x,
          y: origin.y + point.y,
        }))
      : []

    return { ...edge, points } as TRoutedGraphEdge<TEdge>
  })

  return { nodes: positioned, edges: routed }
}
