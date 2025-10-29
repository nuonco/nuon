'use client'

import { useEffect } from 'react'
import {
  ReactFlow,
  Node,
  Edge,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  MarkerType,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'

import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'

const getLayoutedElements = (
  nodes: Node[],
  edges: Edge[],
  direction = 'LR'
) => {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))

  const nodeWidth = 200 // Increased width for longer labels
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
        <Text variant="body" weight="strong">
          App component dependency graph
        </Text>
      }
      triggerButton={{
        children: 'View dependency graph',
        // isMenuButton: true,
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

  const updateNodes = (nodes: any[]) => {
    // First, create a map of nodes with data
    const dataMap = nodes.reduce(
      (acc, node) => {
        if (node.data?.label && node.data?.type) {
          acc[node.id] = {
            label: node.data.label,
            type: node.data.type,
          }
        }
        return acc
      },
      {} as Record<string, { label: string; type: string }>
    )

    // Then update nodes that have empty data
    return nodes.map((node) => {
      if (!node.data?.label && !node.data?.type && dataMap[node.id]) {
        return {
          ...node,
          data: {
            ...node.data,
            label: dataMap[node.id].label,
            type: dataMap[node.id].type,
          },
        }
      }
      return node
    })
  }

  const convertDotToFlowData = (dotGraph: string) => {
    const nodes: Node[] = []
    const edges: Edge[] = []

    // Parse nodes with their attributes
    const nodeRegex = /"([^"]+)"\s*\[\s*([^\]]+)\]/g
    let match

    while ((match = nodeRegex.exec(dotGraph)) !== null) {
      const [, id, attrs] = match
      const attributes = Object.fromEntries(
        attrs.split(',').map((attr) => {
          const [key, value] = attr
            .split('=')
            .map((s) => s.trim().replace(/"/g, ''))
          return [key, value]
        })
      )

      nodes.push({
        id,
        type: 'default',
        data: {
          label: attributes.label,
          type: attributes.type,
        },
        position: { x: 0, y: 0 },
        style: {
          background: attributes.color === 'blue' ? '#1e50c0' : '#991B1B',
          color: '#FAFAFA',
          padding: '8px 12px',
          borderRadius: '4px',
          fontFamily: 'var(--font-hack)',
          fontSize: '12px',
          fontWeight: 500,
          width: 'auto',
          minWidth: '150px',
          textAlign: 'center',
          border: 'none',
        },
      })
    }

    // Parse edges
    const edgeRegex = /"([^"]+)"\s*->\s*"([^"]+)"\s*\[\s*([^\]]+)\]/g
    while ((match = edgeRegex.exec(dotGraph)) !== null) {
      const [, source, target, attrs] = match
      edges.push({
        id: `${source}-${target}`,
        source,
        target,
        type: 'smoothstep',
        animated: false,
        style: {
          stroke: '#991B1B',
          strokeWidth: 2,
        },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: '#991B1B',
        },
      })
    }

    return getLayoutedElements(updateNodes(nodes), edges)
  }

  useEffect(() => {
    if (data) {
      const { nodes: newNodes, edges: newEdges } = convertDotToFlowData(data)
      setNodes(newNodes)
      setEdges(newEdges)
    }
  }, [data])
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
