import { memo, useEffect, useMemo } from 'react'
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

const GraphControls = () => {
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
        onClick={() => fitView({ padding: 0.2 })}
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
  maxZoom?: number
  fitPadding?: number
}

const GraphCanvasInner = ({
  nodes: initialNodes,
  edges: initialEdges,
  nodeTypes,
  height,
  compact,
  maxZoom,
  fitPadding,
}: IGraphCanvas) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)
  const { fitView } = useReactFlow()

  const resolvedMaxZoom = maxZoom ?? (compact ? 1 : 1.5)
  const resolvedPadding = fitPadding ?? (compact ? 0.15 : 0.25)

  useEffect(() => {
    setNodes(initialNodes)
    setEdges(initialEdges)
  }, [initialNodes, initialEdges, setNodes, setEdges])

  const nodeSignature = initialNodes.map((n) => n.id).join('|')
  useEffect(() => {
    const raf = requestAnimationFrame(() =>
      fitView({ padding: resolvedPadding })
    )
    return () => cancelAnimationFrame(raf)
  }, [nodeSignature, fitView, resolvedPadding])

  const memoizedNodeTypes = useMemo(() => nodeTypes, [nodeTypes])

  return (
    <div
      className={cn(
        'relative w-full overflow-hidden border',
        compact ? 'rounded' : 'rounded-lg'
      )}
      style={{ height, background: 'var(--background-neutral)' }}
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoizedNodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
        fitViewOptions={{ padding: resolvedPadding }}
        minZoom={compact ? 0.6 : 0.5}
        maxZoom={resolvedMaxZoom}
        nodesConnectable={false}
        proOptions={{ hideAttribution: true }}
      >
        {!compact && <GraphControls />}
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
