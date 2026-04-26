import { useEffect, useMemo, memo, useState, useCallback } from 'react'
import {
  ReactFlow,
  Node,
  Edge,
  Controls,
  Background,
  MiniMap,
  useNodesState,
  useEdgesState,
  MarkerType,
  Handle,
  Position,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'
import { getSignalGraph } from '@/lib/admin-api'

const NODE_W = 260
const NODE_H = 70

function doLayout(nodes: Node[], edges: Edge[]) {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 30, ranksep: 60 })
  nodes.forEach((n) => {
    const w = n.type === 'signalNode' ? NODE_W : NODE_W - 20
    const h = n.type === 'signalNode' ? NODE_H : NODE_H - 10
    g.setNode(n.id, { width: w, height: h })
  })
  edges.forEach((e) => g.setEdge(e.source, e.target))
  dagre.layout(g)
  return {
    nodes: nodes.map((n) => {
      const p = g.node(n.id)
      const w = n.type === 'signalNode' ? NODE_W : NODE_W - 20
      const h = n.type === 'signalNode' ? NODE_H : NODE_H - 10
      return { ...n, position: { x: p.x - w / 2, y: p.y - h / 2 } }
    }),
    edges,
  }
}

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

function statusColor(s: string): string {
  const l = s.toLowerCase()
  if (l.includes('completed') || l.includes('success')) return '#166534'
  if (l.includes('failed') || l.includes('error')) return '#991B1B'
  if (l.includes('running') || l.includes('executing') || l.includes('active')) return '#1e50c0'
  if (l.includes('pending') || l.includes('queued')) return '#92400E'
  return '#4A545E'
}

const SKIP_NAMES = new Set(['ready', 'Ready'])
function isNoisyUpdate(ue: any) { return SKIP_NAMES.has(ue.name) }

// Build nodes/edges from a single graph node (non-recursive - children handled by expand)
function buildSingleLevel(graphNode: any, parentId: string | null, parentEdgeLabel: string | null, seen: Set<string>): { nodes: Node[]; edges: Edge[] } {
  const nodes: Node[] = []
  const edges: Edge[] = []
  if (!graphNode?.signal) return { nodes, edges }

  const sig = graphNode.signal
  const wfInfo = graphNode.workflow_info
  const status = getStatus(sig.status)
  const id = sig.id
  if (seen.has(id)) return { nodes, edges }
  seen.add(id)

  const updates = (wfInfo?.update_executions || []).filter((ue: any) => !isNoisyUpdate(ue))
  const awaited = wfInfo?.awaited_signals || []
  const childWfs = wfInfo?.child_workflows || []

  // Does this node have children that can be expanded?
  const hasChildren = (graphNode.children?.length > 0) || awaited.length > 0
  const isExpanded = graphNode._expanded

  nodes.push({
    id,
    type: 'signalNode',
    data: {
      signalType: sig.type,
      signalId: sig.id,
      queueId: sig.queue_id,
      status,
      updateCount: updates.length,
      awaitedCount: awaited.length,
      childWfCount: childWfs.length,
      expandable: hasChildren && !isExpanded,
      expanded: isExpanded,
    },
    position: { x: 0, y: 0 },
  })

  if (parentId) {
    edges.push({
      id: `${parentId}->${id}`,
      source: parentId,
      target: id,
      type: 'smoothstep',
      animated: !status.toLowerCase().includes('completed') && !status.toLowerCase().includes('failed'),
      style: { stroke: '#8040BF', strokeWidth: 2 },
      markerEnd: { type: MarkerType.ArrowClosed, color: '#8040BF' },
      label: parentEdgeLabel || 'awaits',
      labelStyle: { fontSize: 9, fill: '#C494F4' },
      labelBgStyle: { fill: '#1B242C', fillOpacity: 0.8 },
    })
  }

  // Chain updates vertically
  for (let i = 0; i < updates.length; i++) {
    const ue = updates[i]
    const ueId = `${id}__ue__${i}`
    nodes.push({
      id: ueId,
      type: 'updateNode',
      data: { name: ue.name, status: ue.status, activityCount: ue.activities?.length || 0, duration: ue.duration },
      position: { x: 0, y: 0 },
    })
    const sourceId = i === 0 ? id : `${id}__ue__${i - 1}`
    edges.push({
      id: `${sourceId}->${ueId}`,
      source: sourceId,
      target: ueId,
      type: 'smoothstep',
      style: { stroke: '#555F6D', strokeWidth: 1.5 },
      markerEnd: { type: MarkerType.ArrowClosed, color: '#555F6D' },
    })
  }

  // Child workflows
  for (const cw of childWfs) {
    const cwId = `${id}__cw__${cw.workflow_id}`
    if (seen.has(cwId)) continue
    seen.add(cwId)
    nodes.push({
      id: cwId,
      type: 'childWfNode',
      data: { workflowType: cw.workflow_type, status: cw.status, namespace: cw.namespace },
      position: { x: 0, y: 0 },
    })
    edges.push({
      id: `${id}->${cwId}`,
      source: id,
      target: cwId,
      type: 'smoothstep',
      style: { stroke: '#1e50c0', strokeWidth: 1.5 },
      markerEnd: { type: MarkerType.ArrowClosed, color: '#1e50c0' },
      label: 'child wf',
      labelStyle: { fontSize: 9, fill: '#6792F4' },
      labelBgStyle: { fill: '#1B242C', fillOpacity: 0.8 },
    })
  }

  // If expanded, recursively add children
  if (graphNode.children) {
    for (const child of graphNode.children) {
      const childResult = buildSingleLevel(child, updates.length > 0 ? `${id}__ue__${updates.length - 1}` : id, child.signal?.type, seen)
      nodes.push(...childResult.nodes)
      edges.push(...childResult.edges)
    }
  }

  return { nodes, edges }
}

// -- Custom Nodes --

const SignalNode = memo(({ data }: any) => {
  const bg = statusColor(data.status)
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ background: '#555' }} />
      <div style={{
        background: bg, color: '#fff', borderRadius: '8px', padding: '10px 14px',
        minWidth: '220px', fontFamily: 'ui-sans-serif, system-ui, sans-serif',
        boxShadow: '0 2px 8px rgba(0,0,0,0.3)',
        cursor: data.expandable ? 'pointer' : 'default',
        border: data.expandable ? '2px dashed rgba(255,255,255,0.4)' : 'none',
      }}>
        <div style={{ fontSize: '12px', fontWeight: 700, fontFamily: 'ui-monospace, monospace', marginBottom: '2px' }}>
          {data.signalType}
        </div>
        <div style={{ fontSize: '9px', opacity: 0.6, fontFamily: 'ui-monospace, monospace', marginBottom: '5px' }}>
          {data.signalId?.slice(0, 20)}
        </div>
        <div style={{ display: 'flex', gap: '6px', fontSize: '10px', flexWrap: 'wrap' }}>
          <span style={{ background: 'rgba(255,255,255,0.2)', borderRadius: '3px', padding: '1px 5px' }}>{data.status}</span>
          {data.updateCount > 0 && <span>{data.updateCount} upd</span>}
          {data.awaitedCount > 0 && <span style={{ color: '#FFD4A8' }}>{data.awaitedCount} await</span>}
          {data.childWfCount > 0 && <span style={{ color: '#8DB0FB' }}>{data.childWfCount} child</span>}
        </div>
        {data.expandable && (
          <div style={{ marginTop: '6px', fontSize: '9px', opacity: 0.7, textAlign: 'center', background: 'rgba(255,255,255,0.1)', borderRadius: '3px', padding: '2px' }}>
            Click to expand
          </div>
        )}
        {data.expanded && (
          <div style={{ marginTop: '6px', fontSize: '9px', opacity: 0.5, textAlign: 'center' }}>
            ✓ expanded
          </div>
        )}
      </div>
      <Handle type="source" position={Position.Bottom} style={{ background: '#555' }} />
    </>
  )
})
SignalNode.displayName = 'SignalNode'

const UpdateNode = memo(({ data }: any) => {
  const border = statusColor(data.status)
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ background: '#555' }} />
      <div style={{
        background: '#272E35', border: `2px solid ${border}`, color: '#fff',
        borderRadius: '6px', padding: '7px 11px', minWidth: '180px',
        fontFamily: 'ui-sans-serif, system-ui, sans-serif',
      }}>
        <div style={{ fontSize: '11px', fontWeight: 600 }}>{data.name}</div>
        <div style={{ display: 'flex', gap: '6px', fontSize: '9px', marginTop: '3px', opacity: 0.6 }}>
          <span>{data.status}</span>
          {data.activityCount > 0 && <span>{data.activityCount} act</span>}
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} style={{ background: '#555' }} />
    </>
  )
})
UpdateNode.displayName = 'UpdateNode'

const ChildWfNode = memo(({ data }: any) => {
  const border = statusColor(data.status)
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ background: '#555' }} />
      <div style={{
        background: '#272E35', border: `2px solid ${border}`, color: '#fff',
        borderRadius: '6px', padding: '7px 11px', minWidth: '180px',
        fontFamily: 'ui-sans-serif, system-ui, sans-serif',
      }}>
        <div style={{ fontSize: '11px', fontWeight: 600 }}>{data.workflowType}</div>
        <div style={{ display: 'flex', gap: '6px', fontSize: '9px', marginTop: '3px', opacity: 0.6 }}>
          <span>{data.status}</span>
          <span>{data.namespace}</span>
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} style={{ background: '#555' }} />
    </>
  )
})
ChildWfNode.displayName = 'ChildWfNode'

const nodeTypes = {
  signalNode: SignalNode,
  updateNode: UpdateNode,
  childWfNode: ChildWfNode,
}

interface ISignalFlowGraph {
  graphData: any
  height?: string
}

export const SignalFlowGraph = ({ graphData, height = '36rem' }: ISignalFlowGraph) => {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [graphTree, setGraphTree] = useState<any>(null)
  const [loading, setLoading] = useState<string | null>(null)

  // Rebuild the visual graph from the tree
  const rebuildGraph = useCallback((tree: any) => {
    if (!tree) return
    const seen = new Set<string>()
    const { nodes: rawNodes, edges: rawEdges } = buildSingleLevel(tree, null, null, seen)
    if (rawNodes.length > 0) {
      const { nodes: ln, edges: le } = doLayout(rawNodes, rawEdges)
      setNodes(ln)
      setEdges(le)
    }
  }, [setNodes, setEdges])

  // Initial load
  useEffect(() => {
    if (graphData && !graphTree) {
      // Mark root as expanded since it comes with workflow_info
      const tree = { ...graphData, _expanded: true }
      setGraphTree(tree)
      rebuildGraph(tree)
    }
  }, [graphData, graphTree, rebuildGraph])

  // Handle click on a signal node to expand it
  const onNodeClick = useCallback(async (_: any, node: Node) => {
    if (node.type !== 'signalNode' || !node.data.expandable) return

    const signalId = node.data.signalId as string
    const queueId = node.data.queueId as string
    if (!signalId || !queueId || loading) return

    setLoading(signalId)
    try {
      const result = await getSignalGraph(queueId, signalId, 1)
      if (result?.graph) {
        // Merge the fetched graph into the tree
        setGraphTree((prev: any) => {
          if (!prev) return prev
          const updated = mergeChildGraph(prev, signalId, result.graph)
          rebuildGraph(updated)
          return updated
        })
      }
    } catch (err) {
      console.error('Failed to expand signal', err)
    } finally {
      setLoading(null)
    }
  }, [loading, rebuildGraph])

  const memoTypes = useMemo(() => nodeTypes, [])

  if (!graphData || nodes.length === 0) return null

  return (
    <div className="w-full border border-gray-200 rounded-lg overflow-hidden relative" style={{ height }}>
      {loading && (
        <div className="absolute top-2 left-2 z-10 bg-gray-900 text-white text-xs px-2 py-1 rounded animate-pulse">
          Loading...
        </div>
      )}
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.05}
        maxZoom={2}
        proOptions={{ hideAttribution: true }}
      >
        <Controls position="top-right" orientation="horizontal" />
        <MiniMap
          nodeColor={(n) => n.type === 'signalNode' ? statusColor(n.data?.status as string || '') : '#272E35'}
          style={{ background: '#0D0D0D' }}
        />
        <Background bgColor="#1B242C" color="#333" gap={20} />
      </ReactFlow>
    </div>
  )
}

// Recursively find a signal node in the tree and merge the expanded graph data into it
function mergeChildGraph(tree: any, targetSignalId: string, childGraph: any): any {
  if (!tree?.signal) return tree

  if (tree.signal.id === targetSignalId) {
    // Found it - merge the workflow info and children, mark as expanded
    return {
      ...tree,
      workflow_info: childGraph.workflow_info || tree.workflow_info,
      children: childGraph.children || [],
      _expanded: true,
    }
  }

  // Recurse into children
  if (tree.children) {
    return {
      ...tree,
      children: tree.children.map((child: any) => mergeChildGraph(child, targetSignalId, childGraph)),
    }
  }

  return tree
}
