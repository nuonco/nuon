import { useMemo } from 'react'
import {
  ReactFlow,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import { Banner } from '@/components/common/Banner'
import {
  StructuralNode,
  ActionNode,
  getLayoutedElements,
  createAddNode,
  createAddEdge,
} from '../plan-graph-utils'

interface IResourceChange {
  urn: string
  type: string
  name: string
  action: string
}

interface IPulumiPlanGraph {
  resources: IResourceChange[]
}

const nodeTypes = {
  structural: StructuralNode,
  action: ActionNode,
}

function buildGraph(resources: IResourceChange[]) {
  const nodes: Node[] = []
  const edges: Edge[] = []
  const nodeIds = new Set<string>()
  const addNode = createAddNode(nodes, nodeIds)
  const addEdge = createAddEdge(edges)

  if (resources.length === 0) return { nodes: [], edges: [] }

  addNode('root', 'structural', { label: 'root' })

  const byType = new Map<string, IResourceChange[]>()
  for (const item of resources) {
    if (!byType.has(item.type)) byType.set(item.type, [])
    byType.get(item.type)!.push(item)
  }

  for (const [type, typeItems] of byType) {
    const typeId = `type-${type}`
    addNode(typeId, 'structural', { label: type })
    addEdge('root', typeId)

    for (const item of typeItems) {
      const itemId = `${item.urn}-${item.action}`
      addNode(itemId, 'action', {
        label: `${item.name} (${item.action})`,
        action: item.action,
      })
      addEdge(typeId, itemId)
    }
  }

  const layouted = getLayoutedElements(nodes, edges)
  return { nodes: layouted, edges }
}

export function PulumiPlanGraph({ resources }: IPulumiPlanGraph) {
  const { nodes: layoutNodes, edges: layoutEdges } = useMemo(
    () => buildGraph(resources),
    [resources],
  )

  const [nodes, setNodes, onNodesChange] = useNodesState(layoutNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(layoutEdges)

  useMemo(() => {
    setNodes(layoutNodes)
    setEdges(layoutEdges)
  }, [layoutNodes, layoutEdges, setNodes, setEdges])

  if (resources.length === 0) {
    return <Banner theme="neutral">No preview data to graph</Banner>
  }

  return (
    <div className="w-full h-[32rem] border rounded-lg bg-white dark:bg-gray-800">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.1}
        maxZoom={1.5}
        nodesDraggable={false}
        proOptions={{ hideAttribution: true }}
        style={{ borderRadius: '8px' }}
      >
        <Controls
          position="top-right"
          orientation="horizontal"
          style={{ color: '#121212' }}
        />
        <Background bgColor="#121212" color="#aaa" gap={16} />
      </ReactFlow>
    </div>
  )
}
