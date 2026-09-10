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

export type TPositionedGraphNode<TNode extends IGraphLayoutNode> = TNode & {
  position: {
    x: number
    y: number
  }
}

const LAYOUT_OPTIONS = {
  algorithm: 'layered',
  hierarchyHandling: 'INCLUDE_CHILDREN',
} as const

const DIRECTION_OPTIONS: Record<TGraphDirection, string> = {
  right: 'RIGHT',
  down: 'DOWN',
}

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
  childrenByParent: Map<string | undefined, IGraphLayoutNode[]>
): ElkNode[] =>
  (childrenByParent.get(parentId) ?? []).map((node) => ({
    id: node.id,
    width: node.width,
    height: node.height,
    children: toElkChildren(node.id, childrenByParent),
  }))

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

export const layoutGraph = async <TNode extends IGraphLayoutNode>({
  nodes,
  edges,
  direction = 'right',
}: IGraphLayoutInput<TNode>): Promise<TPositionedGraphNode<TNode>[]> => {
  const nodesById = new Map(nodes.map((node) => [node.id, node]))
  const parentIds = new Map(
    nodes.map((node) => [node.id, validParentId(node, nodesById)])
  )
  const childrenByParent = new Map<
    string | undefined,
    IGraphLayoutNode[]
  >()

  for (const node of nodes) {
    const parentId = parentIds.get(node.id)
    childrenByParent.set(parentId, [
      ...(childrenByParent.get(parentId) ?? []),
      node,
    ])
  }

  const elk = new ELK()
  const result = await elk.layout({
    id: 'root',
    layoutOptions: {
      'elk.algorithm': LAYOUT_OPTIONS.algorithm,
      'elk.direction': DIRECTION_OPTIONS[direction],
      'elk.hierarchyHandling': LAYOUT_OPTIONS.hierarchyHandling,
    },
    children: toElkChildren(undefined, childrenByParent),
    edges: edges.map(({ id, source, target }) => ({
      id,
      sources: [source],
      targets: [target],
    })),
  })
  const positioned: TPositionedGraphNode<TNode>[] = []

  const appendNodes = (elkNodes: ElkNode[] | undefined) => {
    for (const elkNode of elkNodes ?? []) {
      const node = nodesById.get(elkNode.id)
      if (!node) continue

      const parentId = parentIds.get(node.id)
      positioned.push({
        ...node,
        parentId,
        width: elkNode.width ?? node.width,
        height: elkNode.height ?? node.height,
        position: {
          x: elkNode.x ?? 0,
          y: elkNode.y ?? 0,
        },
      })
      appendNodes(elkNode.children)
    }
  }

  appendNodes(result.children)
  return positioned
}
