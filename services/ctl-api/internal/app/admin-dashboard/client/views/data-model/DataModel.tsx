import { useEffect, useMemo, useState, useCallback, useRef } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getOrgs } from '@/lib/admin-api'
import {
  ReactFlow,
  Node,
  Edge,
  Controls,
  Background,
  MiniMap,
  MarkerType,
  Handle,
  Position,
  ReactFlowInstance,
  useNodesState,
  useEdgesState,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { getDataModel, TCompositeStatus, TDataModelResponse } from '@/lib/admin-api'

type EntityKind =
  | 'org'
  | 'app'
  | 'component'
  | 'install'
  | 'runner'
  | 'queue'
  | 'emitter'
  | 'signal'

const KIND_STYLES: Record<EntityKind, { bg: string; border: string; label: string }> = {
  org:       { bg: '#1e3a8a', border: '#3b82f6', label: 'Org' },
  app:       { bg: '#155e75', border: '#06b6d4', label: 'App' },
  component: { bg: '#134e4a', border: '#14b8a6', label: 'Component' },
  install:   { bg: '#14532d', border: '#22c55e', label: 'Install' },
  runner:    { bg: '#713f12', border: '#eab308', label: 'Runner' },
  queue:     { bg: '#831843', border: '#ec4899', label: 'Queue' },
  emitter:   { bg: '#9d174d', border: '#f472b6', label: 'Emitter' },
  signal:    { bg: '#6b21a8', border: '#c084fc', label: 'Signal' },
}

interface EntityNodeData {
  kind: EntityKind
  name: string
  sub?: string
  href?: string
  status?: TCompositeStatus
  expandable?: boolean
  expanded?: boolean
  onToggle?: (id: string) => void
  id?: string
  [key: string]: unknown
}

const NODE_W = 260
const NODE_H = 76

const EntityNode = ({ data }: { data: EntityNodeData }) => {
  const style = KIND_STYLES[data.kind]
  const statusText = statusLabel(data.status)
  const statusDot = statusColor(data.status)
  return (
    <>
      <Handle type="target" position={Position.Left} style={{ background: '#555' }} isConnectable={false} />
      <div
        style={{
          background: style.bg,
          border: `2px solid ${style.border}`,
          color: '#fff',
          borderRadius: 8,
          padding: '8px 12px',
          width: NODE_W,
          fontFamily: 'ui-sans-serif, system-ui, sans-serif',
          boxShadow: '0 2px 8px rgba(0,0,0,0.35)',
          position: 'relative',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 6 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
            {statusDot && (
              <span
                title={statusText || undefined}
                style={{
                  display: 'inline-block',
                  width: 7,
                  height: 7,
                  borderRadius: '50%',
                  background: statusDot,
                  boxShadow: `0 0 4px ${statusDot}`,
                }}
              />
            )}
            <div style={{ fontSize: 9, opacity: 0.7, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              {style.label}
              {statusText && (
                <span style={{ marginLeft: 6, color: statusDot || '#cbd5e1', opacity: 1, letterSpacing: 0.2 }}>
                  {statusText}
                </span>
              )}
            </div>
          </div>
          {data.href && (
            <a
              href={data.href}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
              onMouseDown={(e) => e.stopPropagation()}
              title="Open detail in new tab"
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: 18,
                height: 18,
                color: '#fff',
                fontSize: 12,
                lineHeight: 1,
                textDecoration: 'none',
                background: 'rgba(255,255,255,0.15)',
                borderRadius: 3,
              }}
            >
              ↗
            </a>
          )}
        </div>
        <div style={{ fontSize: 12, fontWeight: 600, marginTop: 2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {data.name}
        </div>
        {data.sub && (
          <div style={{ fontSize: 9, opacity: 0.75, marginTop: 3, fontFamily: 'ui-monospace, monospace', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {data.sub}
          </div>
        )}
      </div>
      {/* Edge anchor — kept minimal/transparent so React Flow's hardcoded handle
          sizing doesn't fight with our clickable button overlay. */}
      <Handle
        type="source"
        position={Position.Right}
        isConnectable={false}
        style={{
          background: data.expandable ? 'transparent' : '#555',
          border: 'none',
          width: data.expandable ? 1 : 6,
          height: data.expandable ? 1 : 6,
        }}
      />
      {data.expandable && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            if (data.onToggle && data.id) data.onToggle(data.id)
          }}
          onMouseDown={(e) => e.stopPropagation()}
          title={data.expanded ? 'Collapse children' : 'Expand children'}
          style={{
            position: 'absolute',
            top: '50%',
            right: -11,
            transform: 'translateY(-50%)',
            width: 22,
            height: 22,
            padding: 0,
            background: style.border,
            border: `2px solid ${style.bg}`,
            borderRadius: '50%',
            color: '#fff',
            fontSize: 12,
            fontWeight: 700,
            lineHeight: 1,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 2,
          }}
        >
          {data.expanded ? '◂' : '▸'}
        </button>
      )}
    </>
  )
}

const nodeTypes = { entity: EntityNode }

function edge(source: string, target: string, label?: string, color = '#64748b', dashed = false): Edge {
  return {
    id: `${source}->${target}${label ? ':' + label : ''}`,
    source,
    target,
    type: 'smoothstep',
    label,
    labelStyle: { fontSize: 9, fill: '#fff', fontWeight: 600 },
    labelBgStyle: { fill: color, fillOpacity: 0.95 },
    labelBgPadding: [6, 3],
    labelBgBorderRadius: 3,
    style: { stroke: color, strokeWidth: 1.5, strokeDasharray: dashed ? '4 4' : undefined },
    markerEnd: { type: MarkerType.ArrowClosed, color },
  }
}

function makeNode(
  id: string,
  kind: EntityKind,
  name: string,
  sub?: string,
  href?: string,
  status?: TCompositeStatus,
): Node {
  return {
    id,
    type: 'entity',
    position: { x: 0, y: 0 },
    data: { kind, name, sub, href, status },
  }
}

// Polymorphic owner_type strings the BFF emits. Keep these in sync with the
// Go-side TableName() values.
const OWNER_TYPE = {
  org: 'orgs',
  install: 'installs',
  component: 'components',
  runner: 'runners',
} as const

const EDGE_COLOR = {
  product: '#60a5fa',      // org → app → component / install
  runner: '#facc15',       // install ↔ runner
  queue: '#ec4899',        // owner → queue, queue → emitter / signal
} as const

function buildModelGraph(data: TDataModelResponse): { nodes: Node[]; edges: Edge[] } {
  const nodes: Node[] = []
  const edges: Edge[] = []
  if (!data.org) return { nodes, edges }

  // -- Product layer ----------------------------------------------------------
  nodes.push(makeNode(`org:${data.org.id}`, 'org', data.org.name || data.org.id, data.org.id, `/orgs/${data.org.id}`, data.org.status_v2))

  for (const a of data.apps) {
    nodes.push(makeNode(`app:${a.id}`, 'app', a.name || a.id, a.id, undefined, a.status_v2))
    edges.push(edge(`org:${data.org.id}`, `app:${a.id}`, undefined, EDGE_COLOR.product))
  }

  for (const c of data.components) {
    nodes.push(makeNode(`component:${c.id}`, 'component', c.name || c.id, c.type, undefined, c.status_v2))
    edges.push(edge(`app:${c.app_id}`, `component:${c.id}`, undefined, EDGE_COLOR.product))
  }

  for (const i of data.installs) {
    // Install has no CompositeStatus — skip status entirely.
    nodes.push(makeNode(`install:${i.id}`, 'install', i.name || i.id, i.id, `/installs/${i.id}`))
    edges.push(edge(`app:${i.app_id}`, `install:${i.id}`, undefined, EDGE_COLOR.product))
    if (i.runner_id) {
      edges.push(edge(`install:${i.id}`, `runner:${i.runner_id}`, undefined, EDGE_COLOR.runner, true))
    }
  }

  for (const r of data.runners) {
    nodes.push(makeNode(`runner:${r.id}`, 'runner', r.name || r.id, r.id, `/runners/${r.id}`, r.status_v2))
    edges.push(edge(`org:${data.org.id}`, `runner:${r.id}`, undefined, EDGE_COLOR.product))
  }

  const drawn = new Set(nodes.map((n) => n.id))

  // -- Queue layer ------------------------------------------------------------
  // Only draw a queue when we can attach it to its real owner. Queues whose
  // owner_type we don't know how to map (e.g. vcs_connections, runner_processes,
  // general) are intentionally skipped — attaching them anywhere else would lie
  // about the model.
  for (const q of data.queues) {
    const ownerNodeId = ownerKeyFor(q.owner_type, q.owner_id, data.org.id)
    if (!ownerNodeId || !drawn.has(ownerNodeId)) continue
    const qid = `queue:${q.id}`
    nodes.push(
      makeNode(
        qid,
        'queue',
        q.name || 'queue',
        q.max_in_flight ? `max_in_flight=${q.max_in_flight}` : q.owner_type,
        `/queues/${q.id}`,
        q.status_v2,
      ),
    )
    edges.push(edge(ownerNodeId, qid, undefined, EDGE_COLOR.queue))
  }

  const drawnQueues = new Set(data.queues.map((q) => `queue:${q.id}`))
  for (const e of data.emitters) {
    const parent = `queue:${e.queue_id}`
    if (!drawnQueues.has(parent)) continue
    const eid = `emitter:${e.id}`
    nodes.push(makeNode(eid, 'emitter', e.mode || 'emitter', e.id.slice(0, 12), undefined, e.status))
    edges.push(edge(parent, eid, undefined, EDGE_COLOR.queue))
  }

  const drawnEmitters = new Set(data.emitters.map((e) => `emitter:${e.id}`))
  for (const s of data.signals) {
    const parent = `queue:${s.queue_id}`
    if (!drawnQueues.has(parent)) continue
    const sid = `signal:${s.id}`
    nodes.push(
      makeNode(
        sid,
        'signal',
        s.type || 'signal',
        undefined,
        `/queues/${s.queue_id}/signals/${s.id}`,
        s.status,
      ),
    )
    const emitterParent = s.emitter_id ? `emitter:${s.emitter_id}` : null
    if (emitterParent && drawnEmitters.has(emitterParent)) {
      edges.push(edge(emitterParent, sid, undefined, EDGE_COLOR.queue))
    } else {
      edges.push(edge(parent, sid, undefined, EDGE_COLOR.queue))
    }
  }

  return { nodes, edges }
}

function ownerKeyFor(ownerType: string, ownerID: string, orgID: string): string | null {
  switch (ownerType) {
    case OWNER_TYPE.org:
      return ownerID === orgID ? `org:${orgID}` : null
    case OWNER_TYPE.install:
      return `install:${ownerID}`
    case OWNER_TYPE.component:
      return `component:${ownerID}`
    case OWNER_TYPE.runner:
      return `runner:${ownerID}`
    default:
      // Unmapped owner types (vcs_connections, runner_processes, general, ...)
      // are not drawn — adding them later means adding a node kind for each.
      return null
  }
}

function statusLabel(status?: TCompositeStatus): string | undefined {
  return status?.status || undefined
}

// Maps the common Status enum values onto colors. Mirrors the palette in
// SignalFlowGraph.tsx so status semantics stay consistent across admin views.
function statusColor(status?: TCompositeStatus): string | undefined {
  const s = (status?.status || '').toLowerCase()
  if (!s) return undefined
  if (s === 'success' || s.includes('completed') || s === 'active' || s === 'ready') return '#22c55e'
  if (s === 'error' || s === 'failed' || s.includes('failure')) return '#ef4444'
  if (s === 'warning' || s === 'drifted') return '#f59e0b'
  if (s === 'in-progress' || s === 'running' || s === 'planning' || s === 'applying' || s === 'checking-plan' || s === 'provisioning') return '#3b82f6'
  if (s === 'pending' || s === 'queued' || s === 'retrying' || s === 'failed-pending-retry' || s === 'awaiting-user-run') return '#eab308'
  if (s === 'cancelled' || s === 'discarded' || s === 'user-skipped' || s === 'auto-skipped' || s === 'not-attempted' || s === 'expired' || s === 'outdated') return '#94a3b8'
  return '#94a3b8'
}

const RANK_GAP = 80    // horizontal gap between a parent and its children
const SIBLING_GAP = 20 // vertical gap between adjacent siblings

// computeTreeLayout walks a canonical tree from the root and assigns positions
// so that each subtree occupies its own vertical slot. When a node is collapsed
// (or has no visible children) it takes NODE_H of vertical space; when
// expanded, it grows to accommodate its children. Siblings below an expanding
// node naturally shift down to make room — nothing else moves.
function computeTreeLayout(
  rootId: string,
  visible: Set<string>,
  expanded: Set<string>,
  treeChildren: Map<string, string[]>,
): Map<string, { x: number; y: number }> {
  const positions = new Map<string, { x: number; y: number }>()

  // Returns the total vertical height consumed by this node + its visible subtree.
  function walk(nodeId: string, x: number, yStart: number): number {
    const kids = (treeChildren.get(nodeId) || []).filter((c) => visible.has(c))
    if (!expanded.has(nodeId) || kids.length === 0) {
      positions.set(nodeId, { x, y: yStart })
      return NODE_H
    }
    let yCursor = yStart
    for (let i = 0; i < kids.length; i++) {
      if (i > 0) yCursor += SIBLING_GAP
      const h = walk(kids[i], x + NODE_W + RANK_GAP, yCursor)
      yCursor += h
    }
    const totalHeight = yCursor - yStart
    // Center the parent vertically over its children.
    positions.set(nodeId, { x, y: yStart + (totalHeight - NODE_H) / 2 })
    return totalHeight
  }

  walk(rootId, 0, 0)
  return positions
}

const legend: Array<{ kind: EntityKind; what: string }> = [
  { kind: 'org',       what: 'Tenant root' },
  { kind: 'app',       what: 'Application config' },
  { kind: 'component', what: 'Buildable unit' },
  { kind: 'install',   what: 'Deployment of an app' },
  { kind: 'runner',    what: 'Worker in customer VPC' },
  { kind: 'queue',     what: 'Owner-scoped signal queue' },
  { kind: 'emitter',   what: 'Producer attached to a queue' },
  { kind: 'signal',    what: 'Queue item (QueueSignal)' },
]

export const DataModel = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const orgID = searchParams.get('org_id') || ''

  const { data: orgsResp, isLoading: orgsLoading } = useQuery({
    queryKey: ['data-model', 'orgs'],
    queryFn: () => getOrgs({ page: 1 }),
    staleTime: 60_000,
  })
  const orgOptions = useMemo(() => {
    const list = orgsResp?.orgs || []
    return [...list].sort((a, b) => (a.name || a.id).localeCompare(b.name || b.id))
  }, [orgsResp])

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['data-model', orgID],
    queryFn: () => getDataModel(orgID),
    enabled: !!orgID,
  })

  // Full graph + canonical tree. The graph as built has some cross-edges
  // (e.g. install → runner) that aren't tree edges — we pick a canonical
  // parent for each node by BFS from the org, so the tree layout has a single
  // unambiguous parent per node. Non-tree edges still render between their
  // endpoints as additional connections.
  const fullGraph = useMemo(() => {
    if (!data?.org) {
      return {
        nodes: [] as Node[],
        edges: [] as Edge[],
        treeChildren: new Map<string, string[]>(),
        anyChildren: new Map<string, string[]>(),
      }
    }
    const { nodes, edges } = buildModelGraph(data)
    const rootId = `org:${data.org.id}`

    // Adjacency: any outgoing edges (used for "does this node have something
    // to expand?"). Direction follows the source → target edges we built.
    const anyChildren = new Map<string, string[]>()
    for (const e of edges) {
      const arr = anyChildren.get(e.source) || []
      arr.push(e.target)
      anyChildren.set(e.source, arr)
    }

    // Canonical tree: BFS from root, first-visit wins as the parent.
    const treeChildren = new Map<string, string[]>()
    const seen = new Set<string>([rootId])
    const queue = [rootId]
    while (queue.length > 0) {
      const cur = queue.shift()!
      for (const child of anyChildren.get(cur) || []) {
        if (seen.has(child)) continue
        seen.add(child)
        const arr = treeChildren.get(cur) || []
        arr.push(child)
        treeChildren.set(cur, arr)
        queue.push(child)
      }
    }

    return { nodes, edges, treeChildren, anyChildren }
  }, [data])

  // Expanded node IDs. Reset when the org changes; seed with the org so its
  // direct children are visible on first load.
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  useEffect(() => {
    if (data?.org) {
      setExpanded(new Set([`org:${data.org.id}`]))
    } else {
      setExpanded(new Set())
    }
  }, [data?.org?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  const toggleNode = useCallback((id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  // Visible subset: BFS from the org through canonical tree edges only,
  // descending into a node's tree-children if it's expanded.
  const { nodes: laidNodes, edges: laidEdges } = useMemo(() => {
    if (!data?.org || fullGraph.nodes.length === 0) {
      return { nodes: [] as Node[], edges: [] as Edge[] }
    }
    const rootId = `org:${data.org.id}`
    const visible = new Set<string>([rootId])
    const bfsQueue = [rootId]
    while (bfsQueue.length > 0) {
      const cur = bfsQueue.shift()!
      if (!expanded.has(cur)) continue
      for (const child of fullGraph.treeChildren.get(cur) || []) {
        if (!visible.has(child)) {
          visible.add(child)
          bfsQueue.push(child)
        }
      }
    }

    const positions = computeTreeLayout(
      rootId,
      visible,
      expanded,
      fullGraph.treeChildren,
    )

    const visNodes = fullGraph.nodes
      .filter((n) => visible.has(n.id))
      .map((n) => {
        const d = n.data as EntityNodeData
        const hasChildren = (fullGraph.treeChildren.get(n.id) || []).length > 0
        const pos = positions.get(n.id) || { x: 0, y: 0 }
        return {
          ...n,
          position: pos,
          data: {
            ...d,
            id: n.id,
            expandable: hasChildren,
            expanded: expanded.has(n.id),
            onToggle: toggleNode,
          },
        }
      })

    // Render all edges where both endpoints are visible — both tree edges and
    // cross-edges (e.g. install → runner).
    const visEdges = fullGraph.edges.filter(
      (e) => visible.has(e.source) && visible.has(e.target),
    )

    return { nodes: visNodes, edges: visEdges }
  }, [data, fullGraph, expanded, toggleNode])

  const [nodes, setNodes, onNodesChange] = useNodesState(laidNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(laidEdges)
  const memoTypes = useMemo(() => nodeTypes, [])

  const rfRef = useRef<ReactFlowInstance | null>(null)
  const centeredForOrgRef = useRef<string | null>(null)

  useEffect(() => {
    setNodes(laidNodes)
    setEdges(laidEdges)
  }, [laidNodes, laidEdges, setNodes, setEdges])

  // When an org first loads (or we switch orgs), fit the camera to whatever
  // is currently visible — typically the org + its direct children, since
  // the org is auto-expanded on load. We do this exactly once per org so
  // expand/collapse never moves the camera afterwards.
  useEffect(() => {
    if (!data?.org) return
    if (centeredForOrgRef.current === data.org.id) return
    if (!rfRef.current || laidNodes.length < 2) return
    centeredForOrgRef.current = data.org.id
    requestAnimationFrame(() => {
      rfRef.current?.fitView({ padding: 0.2, maxZoom: 1 })
    })
  }, [data?.org?.id, laidNodes])


  const onPickOrg = (next: string) => {
    const params = new URLSearchParams(searchParams)
    if (next) {
      params.set('org_id', next)
    } else {
      params.delete('org_id')
    }
    setSearchParams(params)
  }

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-xl font-semibold text-gray-900 dark:text-gray-100">Data model</h1>
        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Interactive diagram of the Nuon product model (orgs, apps, components, installs,
          runners) and how it maps onto Temporal workflows and Nuon&apos;s signal queue
          primitive. Scoped to a single org. Click a node to expand its children; click
          the ↗ icon to open its detail page in a new tab.
        </p>
      </div>

      <div className="flex items-center gap-2">
        <label htmlFor="data-model-org" className="text-sm text-gray-700 dark:text-gray-300">
          Org
        </label>
        <select
          id="data-model-org"
          value={orgID}
          onChange={(e) => onPickOrg(e.target.value)}
          disabled={orgsLoading}
          className="w-96 rounded border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-1.5 text-sm text-gray-900 dark:text-gray-100"
        >
          <option value="">
            {orgsLoading ? 'Loading orgs…' : 'Select an org…'}
          </option>
          {orgOptions.map((o) => (
            <option key={o.id} value={o.id}>
              {(o.name || o.id) + '  —  ' + o.id}
            </option>
          ))}
        </select>
        {orgsResp && orgsResp.total_pages > 1 && (
          <span className="text-xs text-gray-500 dark:text-gray-400">
            showing first page ({orgOptions.length} of {orgsResp.total_pages * orgOptions.length}+) — use URL ?org_id=… for others
          </span>
        )}
      </div>

      <div
        className="w-full border border-gray-200 rounded-lg overflow-hidden dark:border-gray-800 relative"
        style={{ height: '46rem' }}
      >
        {!orgID && (
          <div className="absolute inset-0 flex items-center justify-center text-sm text-gray-500 dark:text-gray-400">
            Enter an org ID above to load the diagram.
          </div>
        )}
        {orgID && isLoading && (
          <div className="absolute inset-0 flex items-center justify-center text-sm text-gray-500 dark:text-gray-400">
            Loading data model…
          </div>
        )}
        {orgID && isError && (
          <div className="absolute inset-0 flex items-center justify-center text-sm text-red-500">
            Failed to load: {(error as Error)?.message || 'unknown error'}
          </div>
        )}
        {orgID && data && nodes.length === 0 && !isLoading && (
          <div className="absolute inset-0 flex items-center justify-center text-sm text-gray-500 dark:text-gray-400">
            No entities found for this org.
          </div>
        )}
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={memoTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onInit={(instance) => {
            rfRef.current = instance
          }}
          nodesDraggable={false}
          nodesConnectable={false}
          edgesFocusable={false}
          minZoom={0.1}
          maxZoom={2}
          proOptions={{ hideAttribution: true }}
        >
          <Controls position="top-right" orientation="horizontal" />
          <MiniMap
            nodeColor={(node) => {
              const d = node.data as EntityNodeData | undefined
              return d?.kind ? KIND_STYLES[d.kind].border : '#272E35'
            }}
            style={{ background: '#0D0D0D' }}
          />
          <Background bgColor="#1B242C" color="#333" gap={20} />
        </ReactFlow>
      </div>

      <div>
        <h2 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Legend</h2>
        <div className="grid grid-cols-2 lg:grid-cols-3 gap-2">
          {legend.map((l) => {
            const s = KIND_STYLES[l.kind]
            return (
              <div
                key={l.kind}
                className="flex items-center gap-2 rounded border border-gray-200 dark:border-gray-800 px-2 py-1.5 text-xs"
              >
                <span
                  className="inline-block h-3 w-3 rounded"
                  style={{ background: s.bg, border: `1.5px solid ${s.border}` }}
                />
                <span className="font-medium text-gray-800 dark:text-gray-200">{s.label}</span>
                <span className="text-gray-500 dark:text-gray-400 truncate">— {l.what}</span>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
