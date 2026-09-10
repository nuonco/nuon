import {
  BaseEdge,
  Handle,
  Position,
  ReactFlow,
  ReactFlowProvider,
  getSmoothStepPath,
  useEdgesState,
  useNodesState,
  useReactFlow,
  type Edge,
  type EdgeProps,
  type EdgeTypes,
  type Node,
  type NodeProps,
  type NodeTypes,
} from '@xyflow/react'
import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type ComponentType,
  type HTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { useMediaQuery } from '../../../hooks/use-media-query'
import {
  graphLayoutSignature,
  layoutGraph,
  orthogonalEdgePath,
  type IGraphLayoutResult,
  type IGraphPoint,
  type TGraphDirection,
  type TPositionedGraphNode,
  type TRoutedGraphEdge,
} from '../../../utils/graph-layout'
import { Spinner } from '../../atoms/Spinner'
import { Text } from '../../atoms/Text'
import { GraphControls } from './GraphControls'
import '@xyflow/react/dist/style.css'

export const GRAPH_NODE_LIMIT = 80

export interface IGraphNode extends Record<string, unknown> {
  id: string
  kind: string
  parentId?: string
  width: number
  height: number
  data: unknown
  href?: string
}

export interface IGraphEdge {
  id: string
  source: string
  target: string
  kind?: string
}

export interface INodeRenderProps {
  id: string
  kind: string
  parentId?: string
  width: number
  height: number
  data: unknown
  href?: string
  selected: boolean
}

export interface IGraph
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  nodes: IGraphNode[]
  edges: IGraphEdge[]
  nodeTypes: Record<string, ComponentType<INodeRenderProps>>
  direction?: TGraphDirection
  loading?: boolean
  fetching?: boolean
  emptyState?: ReactNode
  height?: number | 'auto'
}

type TGraphFlowNode = Node<IGraphNode>
type TGraphFlowEdge = Edge<{ points: IGraphPoint[] }>
type TLaidOutNode = TPositionedGraphNode<IGraphNode>
type TLaidOutEdge = TRoutedGraphEdge<IGraphEdge>
type TGraphLayout = IGraphLayoutResult<IGraphNode, IGraphEdge>

const DEFAULT_EMPTY = 'No nodes yet'
const DEFAULT_HEIGHT = 480

const targetPosition = (direction: TGraphDirection) =>
  direction === 'down' ? Position.Top : Position.Left

const sourcePosition = (direction: TGraphDirection) =>
  direction === 'down' ? Position.Bottom : Position.Right

const EmptySlot = ({ children }: { children: ReactNode }) => (
  <div className="flex min-h-24 flex-1 items-center justify-center px-4 py-8 text-center">
    {typeof children === 'string' ? (
      <Text color="secondary">{children}</Text>
    ) : (
      children
    )}
  </div>
)

const wrapNodeTypes = (
  nodeTypes: Record<string, ComponentType<INodeRenderProps>>,
  direction: TGraphDirection
): NodeTypes =>
  Object.fromEntries(
    Object.entries(nodeTypes).map(([kind, Renderer]) => [
      kind,
      (props: NodeProps<TGraphFlowNode>) => {
        const node = props.data
        return (
          <>
            <Handle
              type="target"
              position={targetPosition(direction)}
              isConnectable={false}
            />
            <Renderer
              id={node.id}
              kind={node.kind}
              parentId={node.parentId}
              width={props.width ?? node.width}
              height={props.height ?? node.height}
              data={node.data}
              href={node.href}
              selected={props.selected}
            />
            <Handle
              type="source"
              position={sourcePosition(direction)}
              isConnectable={false}
            />
          </>
        )
      },
    ])
  )

const toFlowNodes = (nodes: TLaidOutNode[]): TGraphFlowNode[] =>
  nodes.map((node) => ({
    id: node.id,
    type: node.kind,
    position: node.position,
    data: node,
    width: node.width,
    height: node.height,
    style: { width: node.width, height: node.height },
    parentId: node.parentId,
    extent: node.parentId ? 'parent' : undefined,
    draggable: false,
    connectable: false,
  }))

const RoutedEdge = ({
  data,
  sourceX,
  sourceY,
  sourcePosition,
  targetX,
  targetY,
  targetPosition,
  markerEnd,
}: EdgeProps<TGraphFlowEdge>) => {
  const points = data?.points ?? []
  const [fallback] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  })

  return (
    <BaseEdge
      path={points.length > 1 ? orthogonalEdgePath(points) : fallback}
      markerEnd={markerEnd}
    />
  )
}

const EDGE_TYPES: EdgeTypes = { routed: RoutedEdge }

const toFlowEdges = (
  edges: TLaidOutEdge[],
  selectedIds: Set<string>
): TGraphFlowEdge[] =>
  edges.map((edge) => {
    const dimmed =
      selectedIds.size > 0 &&
      !selectedIds.has(edge.source) &&
      !selectedIds.has(edge.target)

    return {
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: 'routed',
      data: { points: edge.points },
      selectable: false,
      className: dimmed ? 'lite-graph-edge-dimmed' : undefined,
    }
  })

const GraphFrame = ({
  ariaLabel,
  fetching,
  height,
  className,
  children,
  ...props
}: {
  ariaLabel: string
  fetching: boolean
  height: number
  className?: string
  children: ReactNode
} & Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'height'>) => (
  <div
    role="region"
    aria-label={ariaLabel}
    aria-busy={fetching || undefined}
    className={cn(
      'relative flex w-full flex-col overflow-hidden rounded-xl border bg-surface-01',
      className
    )}
    style={{ height }}
    {...props}
  >
    {children}
  </div>
)

const GraphCanvas = ({
  nodes,
  edges,
  nodeTypes,
  direction,
  fetching,
  height,
  nodeCount,
  className,
  ...props
}: {
  nodes: TLaidOutNode[]
  edges: TLaidOutEdge[]
  nodeTypes: Record<string, ComponentType<INodeRenderProps>>
  direction: TGraphDirection
  fetching: boolean
  height: number | 'auto'
  nodeCount: number
  className?: string
} & Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'height'>) => {
  const signature = graphLayoutSignature({ nodes, edges })
  const flowNodeTypes = useMemo(
    () => wrapNodeTypes(nodeTypes, direction),
    [direction, nodeTypes]
  )
  const [flowNodes, setFlowNodes, onNodesChange] =
    useNodesState<TGraphFlowNode>(toFlowNodes(nodes))
  const [flowEdges, setFlowEdges] = useEdgesState<TGraphFlowEdge>(
    toFlowEdges(edges, new Set())
  )
  const { fitView } = useReactFlow()
  const reducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const selectedIds = useMemo(
    () =>
      new Set(flowNodes.filter((node) => node.selected).map((node) => node.id)),
    [flowNodes]
  )
  const selectedKey = [...selectedIds].sort().join('|')

  useEffect(() => {
    setFlowNodes((current) => {
      const selected = new Set(
        current.filter((node) => node.selected).map((node) => node.id)
      )
      return toFlowNodes(nodes).map((node) => ({
        ...node,
        selected: selected.has(node.id),
      }))
    })
  }, [nodes, setFlowNodes])

  useEffect(() => {
    const selected = selectedKey
      ? new Set(selectedKey.split('|'))
      : new Set<string>()
    setFlowEdges(toFlowEdges(edges, selected))
  }, [edges, selectedKey, setFlowEdges])

  useEffect(() => {
    const frame = requestAnimationFrame(() =>
      fitView({ padding: 0.2, duration: reducedMotion ? 0 : 200 })
    )
    return () => cancelAnimationFrame(frame)
  }, [fitView, reducedMotion, signature])

  const layoutHeight = Math.max(
    ...nodes.map((node) =>
      node.parentId ? 0 : node.position.y + node.height
    ),
    0
  )
  const resolvedHeight =
    height === 'auto' ? Math.max(layoutHeight + 48, 240) : height

  return (
    <GraphFrame
      ariaLabel={`Graph with ${nodeCount} nodes`}
      fetching={fetching}
      height={resolvedHeight}
      className={className}
      {...props}
    >
      {fetching ? (
        <span
          role="status"
          aria-label="Refreshing graph"
          className="absolute top-3 left-3 z-10 flex items-center gap-2 text-secondary"
        >
          <Spinner size={14} />
          <Text variant="caption" color="secondary">
            Refreshing
          </Text>
        </span>
      ) : null}
      <ReactFlow
        className="lite-graph h-full"
        nodes={flowNodes}
        edges={flowEdges}
        nodeTypes={flowNodeTypes}
        edgeTypes={EDGE_TYPES}
        onNodesChange={onNodesChange}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.25}
        maxZoom={1.5}
        nodesDraggable={false}
        nodesConnectable={false}
        nodesFocusable={false}
        edgesFocusable={false}
        edgesReconnectable={false}
        elementsSelectable
        panOnDrag
        zoomOnScroll
        proOptions={{ hideAttribution: true }}
        deleteKeyCode={null}
      >
        <GraphControls />
      </ReactFlow>
    </GraphFrame>
  )
}

const GraphLayoutHost = ({
  nodes,
  edges,
  nodeTypes,
  direction,
  loading,
  fetching,
  emptyState = DEFAULT_EMPTY,
  height,
  className,
  ...props
}: IGraph) => {
  const resolvedDirection = direction ?? 'right'
  const tooLarge = nodes.length > GRAPH_NODE_LIMIT
  const empty = !loading && nodes.length === 0
  const signature = graphLayoutSignature({ nodes, edges })
  const inputRef = useRef({ nodes, edges, direction: resolvedDirection })
  inputRef.current = { nodes, edges, direction: resolvedDirection }

  const [layout, setLayout] = useState<TGraphLayout | null>(null)
  const [layoutSignature, setLayoutSignature] = useState<string | null>(null)

  useEffect(() => {
    if (loading || tooLarge || nodes.length === 0) {
      setLayout(null)
      setLayoutSignature(null)
      return
    }

    let cancelled = false
    const input = inputRef.current
    layoutGraph({
      nodes: input.nodes,
      edges: input.edges,
      direction: input.direction,
    }).then((result) => {
      if (cancelled) return
      setLayout(result)
      setLayoutSignature(signature)
    })

    return () => {
      cancelled = true
    }
  }, [loading, nodes.length, signature, tooLarge])

  const laidOut = useMemo(() => {
    if (!layout || layoutSignature !== signature) return null
    const latest = new Map(nodes.map((node) => [node.id, node]))
    return {
      nodes: layout.nodes.map((node) => {
        const current = latest.get(node.id)
        return current
          ? {
              ...current,
              parentId: node.parentId,
              width: node.width,
              height: node.height,
              position: node.position,
            }
          : node
      }),
      edges: layout.edges,
    }
  }, [layout, layoutSignature, nodes, signature])

  const awaitingLayout = !loading && !empty && !tooLarge && !laidOut
  const showLoading = loading || awaitingLayout
  const frameHeight = height === 'auto' ? DEFAULT_HEIGHT : (height ?? DEFAULT_HEIGHT)

  if (tooLarge) {
    return (
      <GraphFrame
        ariaLabel="Graph"
        fetching={false}
        height={frameHeight}
        className={className}
        {...props}
      >
        <EmptySlot>
          <div className="flex max-w-sm flex-col items-center gap-2">
            <Text color="secondary">
              This graph is too large to display. Graphs with more than{' '}
              {GRAPH_NODE_LIMIT} nodes are shown as a list instead.
            </Text>
          </div>
        </EmptySlot>
      </GraphFrame>
    )
  }

  if (showLoading) {
    return (
      <GraphFrame
        ariaLabel="Graph"
        fetching={false}
        height={frameHeight}
        className={className}
        {...props}
      >
        <span role="status" aria-label="Loading graph" className="sr-only">
          Loading graph
        </span>
        <div className="flex flex-1 items-center justify-center">
          <Spinner size={20} />
        </div>
      </GraphFrame>
    )
  }

  if (empty) {
    return (
      <GraphFrame
        ariaLabel="Graph"
        fetching={false}
        height={frameHeight}
        className={className}
        {...props}
      >
        <EmptySlot>{emptyState}</EmptySlot>
      </GraphFrame>
    )
  }

  if (!laidOut) return null

  return (
    <ReactFlowProvider>
      <GraphCanvas
        nodes={laidOut.nodes}
        edges={laidOut.edges}
        nodeTypes={nodeTypes}
        direction={resolvedDirection}
        fetching={fetching}
        height={height ?? DEFAULT_HEIGHT}
        nodeCount={nodes.length}
        className={className}
        {...props}
      />
    </ReactFlowProvider>
  )
}

export const Graph = (props: IGraph) => <GraphLayoutHost {...props} />
