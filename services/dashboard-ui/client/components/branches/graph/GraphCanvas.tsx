import {
  memo,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  type CSSProperties,
} from 'react'
import {
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
  type NodeTypes,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { cn } from '@/utils/classnames'

const GraphControls = ({ onFit }: { onFit?: () => void }) => {
  const { zoomIn, zoomOut, fitView } = useReactFlow()

  return (
    <div
      className="absolute top-3 right-3 z-10 flex items-center gap-1"
      role="toolbar"
      aria-label="Graph controls"
    >
      <Button variant="icon" onClick={() => zoomIn()} aria-label="Zoom in">
        <Icon variant="PlusIcon" size={14} />
      </Button>
      <Button variant="icon" onClick={() => zoomOut()} aria-label="Zoom out">
        <Icon variant="MinusIcon" size={14} />
      </Button>
      <Button
        variant="icon"
        onClick={() => (onFit ? onFit() : fitView({ padding: 0.2 }))}
        aria-label="Fit to view"
      >
        <Icon variant="CornersOutIcon" size={14} />
      </Button>
    </div>
  )
}

interface IGraphCanvas {
  nodes: Node[]
  edges: Edge[]
  nodeTypes: NodeTypes
  height: number
  compact?: boolean
  minZoom?: number
  maxZoom?: number
  fitPadding?: number
  alignStart?: boolean
  style?: CSSProperties
}

const GraphCanvasInner = ({
  nodes: initialNodes,
  edges: initialEdges,
  nodeTypes,
  height,
  compact,
  minZoom,
  maxZoom,
  fitPadding,
  alignStart,
  style,
}: IGraphCanvas) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)
  const { fitView, getNodes, getViewport, setViewport } = useReactFlow()
  const container = useRef<HTMLDivElement>(null)

  const resolvedMaxZoom = maxZoom ?? (compact ? 1 : 1.5)
  const resolvedPadding = fitPadding ?? (compact ? 0.15 : 0.25)
  const fit = useCallback(async () => {
    await fitView({ padding: resolvedPadding })
    if (!alignStart || !container.current) return
    const current = getNodes()
    if (!current.length) return
    const left = Math.min(...current.map((node) => node.position.x))
    const right = Math.max(
      ...current.map(
        (node) => node.position.x + (node.measured?.width ?? node.width ?? 0)
      )
    )
    const top = Math.min(...current.map((node) => node.position.y))
    const bottom = Math.max(
      ...current.map(
        (node) => node.position.y + (node.measured?.height ?? node.height ?? 0)
      )
    )
    const viewport = getViewport()
    await setViewport({
      ...viewport,
      x:
        (right - left) * viewport.zoom > container.current.clientWidth - 48
          ? 24 - left * viewport.zoom
          : viewport.x,
      y:
        (bottom - top) * viewport.zoom > container.current.clientHeight - 96
          ? 64 - top * viewport.zoom
          : viewport.y,
    })
  }, [alignStart, fitView, getNodes, getViewport, setViewport, resolvedPadding])

  useEffect(() => {
    setNodes(initialNodes)
    setEdges(initialEdges)
  }, [initialNodes, initialEdges, setNodes, setEdges])

  const nodeSignature = initialNodes.map((n) => n.id).join('|')
  useEffect(() => {
    let raf = 0
    const refit = () => {
      cancelAnimationFrame(raf)
      raf = requestAnimationFrame(() => {
        void fit()
      })
    }
    refit()
    const observer = alignStart ? new ResizeObserver(refit) : undefined
    if (container.current) observer?.observe(container.current)
    return () => {
      observer?.disconnect()
      cancelAnimationFrame(raf)
    }
  }, [nodeSignature, fit, alignStart])

  const memoizedNodeTypes = useMemo(() => nodeTypes, [nodeTypes])

  return (
    <div
      ref={container}
      className={cn(
        'relative w-full overflow-hidden border',
        compact ? 'rounded' : 'rounded-lg'
      )}
      style={{ height, background: 'var(--background-neutral)', ...style }}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoizedNodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
        fitViewOptions={{ padding: resolvedPadding }}
        minZoom={minZoom ?? (compact ? 0.6 : 0.5)}
        maxZoom={resolvedMaxZoom}
        nodesConnectable={false}
        proOptions={{ hideAttribution: true }}
      >
        {!compact && <GraphControls onFit={alignStart ? fit : undefined} />}
      </ReactFlow>
    </div>
  )
}

export const GraphCanvas = memo((props: IGraphCanvas) => (
  <ReactFlowProvider>
    <GraphCanvasInner {...props} />
  </ReactFlowProvider>
))

GraphCanvas.displayName = 'GraphCanvas'
