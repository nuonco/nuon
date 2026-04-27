import { useState, useMemo, memo, useCallback } from 'react'
import {
  ReactFlow,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  Handle,
  Position,
  MarkerType,
  type Node,
  type Edge,
  type NodeProps,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'

import { Banner } from '@/components/common/Banner'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import type {
  TTerraformResourceChange,
  TTerraformOutputChange,
  TTerraformChangeAction,
} from '@/types'

type Category = 'resources' | 'drift' | 'outputs'

const ACTION_COLORS: Record<TTerraformChangeAction, string> = {
  create: '#16a34a',
  delete: '#dc2626',
  update: '#ea580c',
  replace: '#8040BF',
  read: '#2563eb',
  'no-op': '#6b7280',
}

const STRUCTURAL_COLOR = '#374151'

interface ITerraformPlanGraph {
  resources: TTerraformResourceChange[]
  drift: TTerraformResourceChange[]
  outputs: TTerraformOutputChange[]
}

function estimateNodeWidth(label: string) {
  return Math.max(120, label.length * 8 + 24)
}

function getLayoutedElements(nodes: Node[], edges: Edge[]) {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 40, ranksep: 60 })

  const nodeHeight = 40

  nodes.forEach((node) => {
    const w = estimateNodeWidth(node.data.label as string)
    g.setNode(node.id, { width: w, height: nodeHeight })
  })
  edges.forEach((edge) => {
    g.setEdge(edge.source, edge.target)
  })

  dagre.layout(g)

  const layouted = nodes.map((node) => {
    const w = estimateNodeWidth(node.data.label as string)
    const pos = g.node(node.id)
    return {
      ...node,
      position: { x: pos.x - w / 2, y: pos.y - nodeHeight / 2 },
    }
  })

  return { nodes: layouted, edges }
}

function buildGraph(
  resources: TTerraformResourceChange[],
  drift: TTerraformResourceChange[],
  outputs: TTerraformOutputChange[],
  enabled: Set<Category>,
) {
  const nodes: Node[] = []
  const edges: Edge[] = []
  const nodeIds = new Set<string>()

  const addNode = (id: string, type: string, data: Record<string, unknown>) => {
    if (nodeIds.has(id)) return
    nodeIds.add(id)
    const w = estimateNodeWidth(data.label as string)
    nodes.push({ id, type, data, position: { x: 0, y: 0 }, width: w, style: { background: 'transparent', border: 'none', padding: 0, width: w } })
  }

  const addEdge = (source: string, target: string, isDrift = false) => {
    const color = isDrift ? '#d97706' : '#6b7280'
    edges.push({
      id: `${source}->${target}`,
      source,
      target,
      type: 'smoothstep',
      style: { stroke: color, strokeWidth: 1.5, strokeDasharray: isDrift ? '6 3' : undefined },
      markerEnd: { type: MarkerType.ArrowClosed, color },
    })
  }

  const hasAny =
    (enabled.has('resources') && resources.length > 0) ||
    (enabled.has('drift') && drift.length > 0) ||
    (enabled.has('outputs') && outputs.length > 0)

  if (!hasAny) return { nodes: [], edges: [] }

  addNode('root', 'structural', { label: 'root' })

  const addResourceNodes = (
    items: TTerraformResourceChange[],
    prefix: string,
    isDrift: boolean,
  ) => {
    const byModule = new Map<string, TTerraformResourceChange[]>()
    for (const item of items) {
      const mod = item.module || 'root'
      if (!byModule.has(mod)) byModule.set(mod, [])
      byModule.get(mod)!.push(item)
    }

    for (const [mod, modItems] of byModule) {
      const moduleId = `${prefix}-mod-${mod}`
      if (mod !== 'root') {
        addNode(moduleId, 'structural', { label: mod, isDrift })
        addEdge('root', moduleId, isDrift)
      }
      const parentId = mod === 'root' ? 'root' : moduleId

      const byType = new Map<string, TTerraformResourceChange[]>()
      for (const item of modItems) {
        if (!byType.has(item.resource)) byType.set(item.resource, [])
        byType.get(item.resource)!.push(item)
      }

      for (const [type, typeItems] of byType) {
        const typeId = `${prefix}-type-${mod}-${type}`
        addNode(typeId, 'structural', { label: type, isDrift })
        addEdge(parentId, typeId, isDrift)

        for (const item of typeItems) {
          const itemId = `${prefix}-${item.address}-${item.action}`
          addNode(itemId, 'action', {
            label: `${item.name} (${item.action})`,
            action: item.action,
            isDrift,
          })
          addEdge(typeId, itemId, isDrift)
        }
      }
    }
  }

  if (enabled.has('resources') && resources.length > 0) {
    addResourceNodes(resources, 'res', false)
  }

  if (enabled.has('drift') && drift.length > 0) {
    addResourceNodes(drift, 'drift', true)
  }

  if (enabled.has('outputs') && outputs.length > 0) {
    const outputsGroupId = 'outputs-group'
    addNode(outputsGroupId, 'structural', { label: 'outputs' })
    addEdge('root', outputsGroupId)

    for (const item of outputs) {
      const itemId = `output-${item.output}-${item.action}`
      addNode(itemId, 'output', {
        label: `${item.output} (${item.action})`,
        action: item.action,
      })
      addEdge(outputsGroupId, itemId)
    }
  }

  return getLayoutedElements(nodes, edges)
}

const DRIFT_STRUCTURAL_COLOR = '#92400e'

const StructuralNode = memo(({ data }: NodeProps) => {
  const isDrift = data.isDrift as boolean
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ visibility: 'hidden' }} />
      <div
        className="flex items-center justify-center px-3 py-2"
        style={{
          background: isDrift ? DRIFT_STRUCTURAL_COLOR : STRUCTURAL_COLOR,
          color: '#FAFAFA',
          borderRadius: '6px',
          fontFamily: 'var(--font-hack)',
          fontSize: '12px',
          fontWeight: 600,
          minWidth: '120px',
          textAlign: 'center',
          whiteSpace: 'nowrap',
          border: isDrift ? '2px dashed #d97706' : 'none',
        }}
        title={data.label as string}
      >
        {data.label as string}
      </div>
      <Handle type="source" position={Position.Bottom} style={{ visibility: 'hidden' }} />
    </>
  )
})
StructuralNode.displayName = 'StructuralNode'

const ActionNode = memo(({ data }: NodeProps) => {
  const action = data.action as TTerraformChangeAction
  const isDrift = data.isDrift as boolean
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ visibility: 'hidden' }} />
      <div
        className="flex items-center justify-center px-3 py-2"
        style={{
          background: ACTION_COLORS[action] || '#6b7280',
          color: '#FAFAFA',
          borderRadius: '6px',
          fontFamily: 'var(--font-hack)',
          fontSize: '11px',
          fontWeight: 500,
          minWidth: '120px',
          textAlign: 'center',
          whiteSpace: 'nowrap',
          border: isDrift ? '2px dashed #d97706' : 'none',
        }}
        title={data.label as string}
      >
        {data.label as string}
      </div>
    </>
  )
})
ActionNode.displayName = 'ActionNode'

const OutputNode = memo(({ data }: NodeProps) => {
  const action = data.action as TTerraformChangeAction
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ visibility: 'hidden' }} />
      <div
        className="flex items-center justify-center px-3 py-2"
        style={{
          background: ACTION_COLORS[action] || '#6b7280',
          color: '#FAFAFA',
          borderRadius: '6px',
          fontFamily: 'var(--font-hack)',
          fontSize: '11px',
          fontWeight: 500,
          minWidth: '120px',
          textAlign: 'center',
          whiteSpace: 'nowrap',
        }}
        title={data.label as string}
      >
        {data.label as string}
      </div>
    </>
  )
})
OutputNode.displayName = 'OutputNode'

const nodeTypes = {
  structural: StructuralNode,
  action: ActionNode,
  output: OutputNode,
}

export function TerraformPlanGraph({
  resources,
  drift,
  outputs,
}: ITerraformPlanGraph) {
  const [enabled, setEnabled] = useState<Set<Category>>(
    new Set(['resources', 'drift', 'outputs']),
  )

  const toggle = useCallback((cat: Category) => {
    setEnabled((prev) => {
      const next = new Set(prev)
      if (next.has(cat)) next.delete(cat)
      else next.add(cat)
      return next
    })
  }, [])

  const { nodes: layoutNodes, edges: layoutEdges } = useMemo(
    () => buildGraph(resources, drift, outputs, enabled),
    [resources, drift, outputs, enabled],
  )

  const [nodes, setNodes, onNodesChange] = useNodesState(layoutNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(layoutEdges)

  useMemo(() => {
    setNodes(layoutNodes)
    setEdges(layoutEdges)
  }, [layoutNodes, layoutEdges, setNodes, setEdges])

  const isEmpty =
    resources.length === 0 && drift.length === 0 && outputs.length === 0

  if (isEmpty) {
    return <Banner theme="neutral">No plan data to graph</Banner>
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex gap-4">
        <CheckboxInput
          checked={enabled.has('resources')}
          onChange={() => toggle('resources')}
          labelProps={{ labelText: `Resources (${resources.length})` }}
        />
        <CheckboxInput
          checked={enabled.has('drift')}
          onChange={() => toggle('drift')}
          labelProps={{ labelText: `Drift (${drift.length})` }}
        />
        <CheckboxInput
          checked={enabled.has('outputs')}
          onChange={() => toggle('outputs')}
          labelProps={{ labelText: `Outputs (${outputs.length})` }}
        />
      </div>

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
    </div>
  )
}
