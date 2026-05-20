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

const NODE_W = 280
const NODE_H = 90

function doLayout(nodes: Node[], edges: Edge[]) {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 30, ranksep: 60 })
  nodes.forEach((n) => {
    const w = n.type === 'signalNode' ? NODE_W : NODE_W - 30
    const h = n.type === 'signalNode' ? NODE_H : 60
    g.setNode(n.id, { width: w, height: h })
  })
  edges.forEach((e) => g.setEdge(e.source, e.target))
  dagre.layout(g)
  return {
    nodes: nodes.map((n) => {
      const p = g.node(n.id)
      const w = n.type === 'signalNode' ? NODE_W : NODE_W - 30
      const h = n.type === 'signalNode' ? NODE_H : 60
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

function buildGraph(graphNode: any, parentId: string | null, label: string | null, expandedSet: Set<string>, seen: Set<string>): { nodes: Node[]; edges: Edge[] } {
  const nodes: Node[] = []
  const edges: Edge[] = []
  if (!graphNode?.signal) return { nodes, edges }

  const sig = graphNode.signal
  const wfInfo = graphNode.workflow_info
  const status = getStatus(sig.status)
  const id = sig.id
  if (seen.has(id)) return { nodes, edges }
  seen.add(id)

  const updates = (wfInfo?.update_executions || []).filter((ue: any) => !SKIP_NAMES.has(ue.name))
  const awaited = wfInfo?.awaited_signals || []
  const enqueued = wfInfo?.enqueued_signals || []
  const isExpanded = expandedSet.has(id)
  const hasWfInfo = !!wfInfo

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
      enqueuedCount: enqueued.length,
      expanded: isExpanded,
      hasWfInfo,
    },
    position: { x: 0, y: 0 },
  })

  if (parentId) {
    const isEnqueued = label === 'enqueued'
    const edgeColor = isEnqueued ? '#22C55E' : '#8040BF'
    const labelColor = isEnqueued ? '#86EFAC' : '#C494F4'
    edges.push({
      id: `${parentId}->${id}`,
      source: parentId,
      target: id,
      type: 'smoothstep',
      animated: !status.toLowerCase().includes('completed') && !status.toLowerCase().includes('failed'),
      style: { stroke: edgeColor, strokeWidth: 2 },
      markerEnd: { type: MarkerType.ArrowClosed, color: edgeColor },
      label: label || undefined,
      labelStyle: { fontSize: 9, fill: labelColor },
      labelBgStyle: { fill: '#1B242C', fillOpacity: 0.8 },
    })
  }

  // Only show updates + children if this node is expanded
  if (!isExpanded) return { nodes, edges }

  // Updates
  for (let i = 0; i < updates.length; i++) {
    const ue = updates[i]
    const ueId = `${id}__ue__${i}`
    nodes.push({
      id: ueId,
      type: 'updateNode',
      data: {
        name: ue.name,
        status: ue.status,
        activityCount: ue.activities?.length || 0,
        activities: ue.activities || [],
        input: ue.input,
        result: ue.result,
        failure: ue.failure,
      },
      position: { x: 0, y: 0 },
    })
    const src = i === 0 ? id : `${id}__ue__${i - 1}`
    edges.push({
      id: `${src}->${ueId}`,
      source: src,
      target: ueId,
      type: 'smoothstep',
      style: { stroke: '#555F6D', strokeWidth: 1.5 },
      markerEnd: { type: MarkerType.ArrowClosed, color: '#555F6D' },
    })
  }

  // Recurse into children
  if (graphNode.children) {
    for (const child of graphNode.children) {
      const lastUpdate = updates.length > 0 ? `${id}__ue__${updates.length - 1}` : id
      const edgeLabel = child.relationship || child.signal?.type
      const sub = buildGraph(child, lastUpdate, edgeLabel, expandedSet, seen)
      nodes.push(...sub.nodes)
      edges.push(...sub.edges)
    }
  }

  return { nodes, edges }
}

// -- Node components with action buttons --

const SignalNode = memo(({ data }: any) => {
  const bg = statusColor(data.status)
  return (
    <>
      <Handle type="target" position={Position.Top} style={{ background: '#555' }} />
      <div style={{
        background: bg, color: '#fff', borderRadius: '8px', padding: '10px 14px',
        width: `${NODE_W}px`, fontFamily: 'ui-sans-serif, system-ui, sans-serif',
        boxShadow: '0 2px 8px rgba(0,0,0,0.3)',
      }}>
        <div style={{ fontSize: '12px', fontWeight: 700, fontFamily: 'ui-monospace, monospace', marginBottom: '2px' }}>
          {data.signalType}
        </div>
        <div style={{ fontSize: '9px', opacity: 0.6, fontFamily: 'ui-monospace, monospace', marginBottom: '4px' }}>
          {data.signalId?.slice(0, 22)}
        </div>
        <div style={{ display: 'flex', gap: '6px', fontSize: '10px', flexWrap: 'wrap', marginBottom: '6px' }}>
          <span style={{ background: 'rgba(255,255,255,0.2)', borderRadius: '3px', padding: '1px 5px' }}>{data.status}</span>
          {data.updateCount > 0 && <span>{data.updateCount} upd</span>}
          {data.awaitedCount > 0 && <span style={{ color: '#FFD4A8' }}>{data.awaitedCount} await</span>}
          {data.enqueuedCount > 0 && <span style={{ color: '#75CC9E' }}>{data.enqueuedCount} enq</span>}
        </div>
        <div style={{ display: 'flex', gap: '4px' }}>
          <button
            className="signal-graph-btn"
            data-action="view-signal"
            data-signal-id={data.signalId}
            data-queue-id={data.queueId}
            style={{
              background: 'rgba(255,255,255,0.2)', border: 'none', borderRadius: '3px',
              color: '#fff', fontSize: '9px', padding: '2px 8px', cursor: 'pointer',
            }}
          >
            View signal
          </button>
          {!data.expanded && (
            <button
              className="signal-graph-btn"
              data-action="expand"
              data-signal-id={data.signalId}
              data-queue-id={data.queueId}
              style={{
                background: 'rgba(255,255,255,0.3)', border: '1px dashed rgba(255,255,255,0.5)',
                borderRadius: '3px', color: '#fff', fontSize: '9px', padding: '2px 8px', cursor: 'pointer',
              }}
            >
              ▸ Expand
            </button>
          )}
          {data.expanded && (
            <span style={{ fontSize: '9px', opacity: 0.5, padding: '2px 4px' }}>✓ expanded</span>
          )}
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} style={{ background: '#555' }} />
    </>
  )
})
SignalNode.displayName = 'SignalNode'

const UpdateNodeInner = ({ data }: any) => {
  const [showAll, setShowAll] = useState(false)
  const border = statusColor(data.status)
  const acts: any[] = data.activities || []
  const visible = showAll ? acts : acts.slice(0, 3)

  return (
    <>
      <Handle type="target" position={Position.Top} style={{ background: '#555' }} />
      <div style={{
        background: '#272E35', border: `2px solid ${border}`, color: '#fff',
        borderRadius: '6px', padding: '7px 11px', width: `${NODE_W - 30}px`,
        fontFamily: 'ui-sans-serif, system-ui, sans-serif',
      }}>
        <div style={{ fontSize: '11px', fontWeight: 600 }}>{data.name}</div>
        <div style={{ display: 'flex', gap: '6px', fontSize: '9px', marginTop: '3px', opacity: 0.7 }}>
          <span>{data.status}</span>
          {acts.length > 0 && <span>{acts.length} activities</span>}
        </div>
        {data.failure && (
          <div style={{ fontSize: '9px', color: '#FCA5A5', marginTop: '4px', maxHeight: '24px', overflow: 'hidden' }}>
            ✗ {data.failure.slice(0, 80)}{data.failure.length > 80 ? '...' : ''}
          </div>
        )}
        {acts.length > 0 && (
          <div style={{ marginTop: '5px', borderTop: '1px solid rgba(255,255,255,0.1)', paddingTop: '4px' }}>
            {visible.map((a: any, i: number) => (
              <div key={i} style={{ fontSize: '8px', display: 'flex', gap: '4px', opacity: 0.8, marginBottom: '2px', alignItems: 'center' }}>
                <span style={{ color: a.status === 'Completed' ? '#75CC9E' : a.status === 'Failed' ? '#FCA5A5' : a.status === 'Running' ? '#6792F4' : '#9EA8B3', flexShrink: 0 }}>●</span>
                <span style={{ fontFamily: 'ui-monospace, monospace', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: '8px' }}>{a.name}</span>
                <span style={{ fontSize: '7px', opacity: 0.5, flexShrink: 0 }}>{a.status}</span>
              </div>
            ))}
            {acts.length > 3 && (
              <button
                onClick={(e) => { e.stopPropagation(); setShowAll(!showAll) }}
                style={{
                  fontSize: '8px', color: '#AD71EA', background: 'rgba(128,64,191,0.15)',
                  border: '1px solid rgba(128,64,191,0.3)', borderRadius: '3px',
                  padding: '2px 6px', cursor: 'pointer', marginTop: '3px', width: '100%',
                }}
              >
                {showAll ? 'Show less' : `Show all ${acts.length} activities`}
              </button>
            )}
          </div>
        )}
      </div>
      <Handle type="source" position={Position.Bottom} style={{ background: '#555' }} />
    </>
  )
}
const UpdateNode = memo(UpdateNodeInner)
UpdateNode.displayName = 'UpdateNode'

const nodeTypes = { signalNode: SignalNode, updateNode: UpdateNode }

interface ISignalFlowGraph {
  graphData: any
  height?: string
}

export const SignalFlowGraph = ({ graphData, height = '36rem' }: ISignalFlowGraph) => {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [graphTree, setGraphTree] = useState<any>(null)
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState<string | null>(null)

  const rebuild = useCallback((tree: any, expanded: Set<string>) => {
    if (!tree) return
    const seen = new Set<string>()
    const { nodes: rn, edges: re } = buildGraph(tree, null, null, expanded, seen)
    if (rn.length > 0) {
      const { nodes: ln, edges: le } = doLayout(rn, re)
      setNodes(ln)
      setEdges(le)
    }
  }, [setNodes, setEdges])

  // Initial load - auto-expand root
  useEffect(() => {
    if (graphData && !graphTree) {
      const tree = graphData
      setGraphTree(tree)
      const rootId = tree.signal?.id
      const initial = new Set<string>()
      if (rootId) initial.add(rootId)
      setExpandedIds(initial)
      rebuild(tree, initial)
    }
  }, [graphData, graphTree, rebuild])

  // Handle button clicks inside nodes via event delegation
  useEffect(() => {
    const handler = async (e: MouseEvent) => {
      const btn = (e.target as HTMLElement).closest('.signal-graph-btn') as HTMLElement | null
      if (!btn) return

      const action = btn.dataset.action
      const signalId = btn.dataset.signalId
      const queueId = btn.dataset.queueId

      if (action === 'view-signal' && signalId && queueId) {
        window.open(`/queues/${queueId}/signals/${signalId}`, '_blank')
        return
      }

      if (action === 'expand' && signalId && queueId && !loading) {
        setLoading(signalId)
        try {
          const result = await getSignalGraph(queueId, signalId, 1)
          if (result?.graph) {
            setGraphTree((prev: any) => {
              const updated = mergeChild(prev, signalId, result.graph)
              const newExpanded = new Set(expandedIds)
              newExpanded.add(signalId)
              setExpandedIds(newExpanded)
              rebuild(updated, newExpanded)
              return updated
            })
          }
        } catch (err) {
          console.error('Failed to expand signal', err)
        } finally {
          setLoading(null)
        }
      }
    }

    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [loading, expandedIds, rebuild])

  const memoTypes = useMemo(() => nodeTypes, [])

  if (!graphData || nodes.length === 0) return null

  return (
    <div className="w-full border border-gray-200 rounded-lg overflow-hidden relative dark:border-gray-800" style={{ height }}>
      {loading && (
        <div className="absolute top-2 left-2 z-10 bg-gray-900 text-white text-xs px-3 py-1.5 rounded shadow-lg animate-pulse">
          Expanding signal...
        </div>
      )}
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={memoTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.02}
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

function mergeChild(tree: any, targetId: string, childGraph: any): any {
  if (!tree?.signal) return tree
  if (tree.signal.id === targetId) {
    return {
      ...tree,
      workflow_info: childGraph.workflow_info || tree.workflow_info,
      children: childGraph.children || tree.children || [],
    }
  }
  if (tree.children) {
    return { ...tree, children: tree.children.map((c: any) => mergeChild(c, targetId, childGraph)) }
  }
  return tree
}
