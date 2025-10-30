'use client'

import { useEffect, memo, useMemo } from 'react'
import {
  ReactFlow,
  Node,
  Edge,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  MarkerType,
  NodeProps,
  Handle,
  Position,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'

import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { Modal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TComponentType } from '@/types'

const getLayoutedElements = (
  nodes: Node[],
  edges: Edge[],
  direction = 'LR'
) => {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))

  const nodeWidth = 200
  const nodeHeight = 40

  dagreGraph.setGraph({ rankdir: direction })

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight })
  })

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target)
  })

  dagre.layout(dagreGraph)

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id)
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    }
  })

  return { nodes: layoutedNodes, edges }
}

export const ComponentsGraphRenderer = ({
  appId,
  configId,
}: {
  appId: string
  configId: string
}) => {
  return (
    <Modal
      heading={
        <Text
          className="!inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="GraphIcon" size="24" />
          Component dependency graph
        </Text>
      }
      triggerButton={{
        children: (
          <>
            View dependency graph <Icon variant="GraphIcon" />
          </>
        ),
        isMenuButton: true,
        variant: 'ghost',
      }}
      size="full"
    >
      <div className="flex flex-col gap-2">
        <Text>
          Nuon automatically creates a graph of all of the components in your
          application.
        </Text>

        <ul className="flex flex-col gap-1 list-disc pl-4">
          <li className="text-sm max-w-xl">
            Dependencies are from root to dependencies (so a red-arrow from a to
            b, means that b depends on a, or that when a changes, b would be
            updated when <Code variant="inline">select-dependencies</Code> is
            true)
          </li>
          <li className="text-sm">
            Blue nodes mean that the current config version has changes to that
            component
          </li>
        </ul>
      </div>
      <ComponentsGraph appId={appId} configId={configId} />
    </Modal>
  )
}

// Custom node component that renders the JSX
const CustomComponentNode = memo(({ data, id }: NodeProps) => {
  const backgroundColor = data.color === 'blue' ? '#1e50c0' : '#991B1B'

  return (
    <>
      <Handle type="target" position={Position.Left} />
      <div
        className="flex items-center gap-2 px-3 py-2"
        style={{
          background: backgroundColor,
          color: '#FAFAFA',
          borderRadius: '4px',
          fontFamily: 'var(--font-hack)',
          fontSize: '12px',
          fontWeight: 500,
          minWidth: '150px',
          textAlign: 'center',
          border: 'none',
          whiteSpace: 'nowrap',
        }}
      >
        {data.componentType && (
          <ComponentType
            type={data.componentType as TComponentType}
            displayVariant="icon-only"
            variant="subtext"
          />
        )}
        <span>{(data.componentLabel as string) || id}</span>
      </div>
      <Handle type="source" position={Position.Right} />
    </>
  )
})

CustomComponentNode.displayName = 'CustomComponentNode'

const nodeTypes = {
  customComponent: CustomComponentNode,
}

const ComponentsGraph = ({
  appId,
  configId,
}: {
  appId: string
  configId: string
}) => {
  const { org } = useOrg()
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

  const { data, error, isLoading } = useQuery({
    path: `/api/orgs/${org?.id}/apps/${appId}/configs/${configId}/graph`,
  })

  const convertDotToFlowData = (dotGraph: string) => {
    const nodesMap = new Map<string, Node>()
    const edges: Edge[] = []
    const allNodeIds = new Set<string>()

    // First pass: Find all node declarations with attributes
    // Must NOT have "->" before the node (which would make it part of an edge)
    const nodeWithAttrsRegex = /^\s*"([^"]+)"\s*\[\s*([^\]]+?)\s*\];?\s*$/gm
    let match

    while ((match = nodeWithAttrsRegex.exec(dotGraph)) !== null) {
      const [fullMatch, id, attrs] = match

      allNodeIds.add(id)

      // Parse attributes
      const attributes: Record<string, string> = {}
      const attrRegex = /(\w+)\s*=\s*"([^"]*)"/g
      let attrMatch

      while ((attrMatch = attrRegex.exec(attrs)) !== null) {
        attributes[attrMatch[1]] = attrMatch[2]
      }

      const nodeData = {
        componentLabel: String(attributes.label || attributes.name || id),
        componentType: String(attributes.type || '') as TComponentType,
        color: attributes.color === 'blue' ? 'blue' : 'red',
      }

      nodesMap.set(id, {
        id: String(id),
        type: 'customComponent',
        data: nodeData,
        position: { x: 0, y: 0 },
      })
    }

    // Second pass: Parse edges and collect any node IDs not yet seen
    const edgeRegex =
      /^\s*"([^"]+)"\s*->\s*"([^"]+)"\s*\[\s*([^\]]*)\s*\];?\s*$/gm
    while ((match = edgeRegex.exec(dotGraph)) !== null) {
      const [, source, target, attrs] = match

      // Track node IDs
      allNodeIds.add(source)
      allNodeIds.add(target)

      // Parse edge attributes
      const attributes: Record<string, string> = {}
      const attrRegex = /(\w+)\s*=\s*"([^"]*)"/g
      let attrMatch

      while ((attrMatch = attrRegex.exec(attrs)) !== null) {
        attributes[attrMatch[1]] = attrMatch[2]
      }

      const edgeColor = attributes.color === 'red' ? '#991B1B' : '#1e50c0'

      edges.push({
        id: `${source}-${target}`,
        source: String(source),
        target: String(target),
        type: 'smoothstep',
        animated: false,
        style: {
          stroke: edgeColor,
          strokeWidth: 2,
        },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: edgeColor,
        },
      })
    }

    // Third pass: Create basic nodes for any IDs that don't have attribute declarations
    allNodeIds.forEach((id) => {
      if (!nodesMap.has(id)) {
        nodesMap.set(id, {
          id: String(id),
          type: 'customComponent',
          data: {
            componentLabel: String(id),
            componentType: '' as TComponentType,
            color: 'red',
          },
          position: { x: 0, y: 0 },
        })
      }
    })

    const nodes = Array.from(nodesMap.values())

    return getLayoutedElements(nodes, edges)
  }

  useEffect(() => {
    if (data) {
      const { nodes: newNodes, edges: newEdges } = convertDotToFlowData(data)
      setNodes(newNodes)
      setEdges(newEdges)
    }
  }, [data, setNodes, setEdges])

  const memoizedNodeTypes = useMemo(() => nodeTypes, [])

  return (
    <>
      {isLoading ? (
        <Skeleton width="100%" height="32rem" />
      ) : error?.error ? (
        <Banner theme="error">
          {error?.error || 'Unable to load component change graph.'}
        </Banner>
      ) : (
        <div className="w-full h-[32rem] border rounded-lg bg-white dark:bg-gray-800">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={memoizedNodeTypes}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            fitView
            fitViewOptions={{ padding: 0.2 }}
            minZoom={0.1}
            maxZoom={1.5}
            defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
            proOptions={{ hideAttribution: true }}
            style={{
              borderRadius: '8px',
            }}
          >
            <Controls
              position="top-right"
              orientation="horizontal"
              style={{
                color: '#121212',
              }}
            />
            <Background bgColor="#121212" color="#aaa" gap={16} />
          </ReactFlow>
        </div>
      )}
    </>
  )
}
