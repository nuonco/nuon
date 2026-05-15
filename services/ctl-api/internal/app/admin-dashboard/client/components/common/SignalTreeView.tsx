import { useState, useCallback } from 'react'
import { Link } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { getSignalGraph } from '@/lib/admin-api'
import { truncateId, formatDuration } from '@/utils/format'

interface SignalGraphNode {
  signal?: any
  workflow_info?: any
  children?: SignalGraphNode[]
  relationship?: string
}

interface ISignalTreeView {
  graphData: SignalGraphNode
  temporalUIUrl?: string
  height?: string
}

export const SignalTreeView = ({ graphData, temporalUIUrl, height = '36rem' }: ISignalTreeView) => {
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())
  const [mergedGraph, setMergedGraph] = useState<SignalGraphNode>(graphData)
  const [expandedRemote, setExpandedRemote] = useState<Set<string>>(new Set())
  const [loadingId, setLoadingId] = useState<string | null>(null)

  // Sync when graphData prop changes (initial load or parent re-fetch)
  if (graphData !== mergedGraph && expandedRemote.size === 0) {
    setMergedGraph(graphData)
  }

  const toggleCollapsed = useCallback((id: string) => {
    setCollapsed(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }, [])

  const handleExpand = useCallback(async (queueId: string, signalId: string) => {
    if (loadingId) return
    setLoadingId(signalId)
    try {
      const result = await getSignalGraph(queueId, signalId, 1)
      if (result?.graph) {
        setMergedGraph(prev => mergeChild(prev, signalId, result.graph))
        setExpandedRemote(prev => new Set(prev).add(signalId))
      }
    } catch (err) {
      console.error('Failed to expand signal', err)
    } finally {
      setLoadingId(null)
    }
  }, [loadingId])

  const collapseAll = useCallback(() => {
    const ids = new Set<string>()
    const walk = (node: SignalGraphNode) => {
      if (node.signal?.id) ids.add(node.signal.id)
      node.children?.forEach(walk)
    }
    walk(mergedGraph)
    setCollapsed(ids)
  }, [mergedGraph])

  const expandAll = useCallback(() => setCollapsed(new Set()), [])

  const isAllExpanded = collapsed.size === 0

  return (
    <div className="w-full border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden" style={{ height, maxHeight: height }}>
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
        <span className="text-xs font-medium text-gray-700 dark:text-gray-300">Signal tree</span>
        <div className="flex items-center gap-2">
          {loadingId && (
            <span className="text-[10px] text-primary-500 animate-pulse">Loading...</span>
          )}
          <button
            onClick={isAllExpanded ? collapseAll : expandAll}
            className="text-[10px] text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 px-1.5 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-800"
          >
            {isAllExpanded ? 'Collapse all' : 'Expand all'}
          </button>
        </div>
      </div>

      {/* Tree body */}
      <div className="overflow-y-auto" style={{ height: `calc(${height} - 2.25rem)` }}>
        <TreeNode
          node={mergedGraph}
          depth={0}
          collapsed={collapsed}
          onToggle={toggleCollapsed}
          onExpand={handleExpand}
          loadingId={loadingId}
          expandedRemote={expandedRemote}
          temporalUIUrl={temporalUIUrl}
        />
      </div>
    </div>
  )
}

interface ITreeNode {
  node: SignalGraphNode
  depth: number
  collapsed: Set<string>
  onToggle: (id: string) => void
  onExpand: (queueId: string, signalId: string) => void
  loadingId: string | null
  expandedRemote: Set<string>
  temporalUIUrl?: string
}

const TreeNode = ({
  node,
  depth,
  collapsed,
  onToggle,
  onExpand,
  loadingId,
  expandedRemote,
  temporalUIUrl,
}: ITreeNode) => {
  const signal = node.signal
  if (!signal?.id) return null

  const hasChildren = (node.children?.length ?? 0) > 0
  const hasWorkflowInfo = !!node.workflow_info
  const isCollapsed = collapsed.has(signal.id)
  const canExpandRemote = !hasChildren && hasWorkflowInfo && !expandedRemote.has(signal.id)
  const isLoading = loadingId === signal.id
  const status = getStatus(signal.status)
  const statusStr = statusString(status)
  const dotColor = statusDotColor(status)
  const relationshipColor = node.relationship === 'enqueued'
    ? 'text-green-500 dark:text-green-400'
    : node.relationship === 'awaited'
      ? 'text-orange-500 dark:text-orange-400'
      : ''

  return (
    <>
      <div className="group flex items-stretch min-h-[28px] hover:bg-gray-50 dark:hover:bg-gray-800/50">
        {/* Depth indentation lines */}
        {Array.from({ length: depth }).map((_, i) => (
          <span key={i} className="w-4 shrink-0 flex justify-center">
            <span className="w-px self-stretch bg-gray-200 dark:bg-gray-700" />
          </span>
        ))}

        <div className="flex items-center gap-1.5 flex-1 min-w-0 px-2 py-1">
          {/* Expand/collapse caret */}
          {hasChildren ? (
            <button
              type="button"
              className="flex items-center justify-center w-4 h-4 shrink-0 rounded text-gray-400 dark:text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-700"
              onClick={() => onToggle(signal.id)}
            >
              <span className="text-[10px]">{isCollapsed ? '▸' : '▾'}</span>
            </button>
          ) : (
            <span className="w-4 shrink-0" />
          )}

          {/* Status dot */}
          <span className={`${dotColor} text-[8px] shrink-0`}>●</span>

          {/* Signal type */}
          <span className="font-mono text-xs text-gray-900 dark:text-gray-100 truncate shrink-0">
            {signal.type}
          </span>

          {/* Signal ID link */}
          <Link
            to={`/queues/${signal.queue_id}/signals/${signal.id}`}
            className="font-mono text-[10px] text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 shrink-0"
          >
            {truncateId(signal.id)}
          </Link>

          {/* Relationship badge */}
          {node.relationship && (
            <span className={`text-[9px] font-medium ${relationshipColor} shrink-0`}>
              {node.relationship}
            </span>
          )}

          {/* Status badge */}
          <Badge variant="status" status={statusStr}>{statusStr}</Badge>

          {/* Workflow link */}
          {signal.workflow?.id && signal.workflow?.namespace && temporalUIUrl && (
            <a
              href={`${temporalUIUrl}/namespaces/${signal.workflow.namespace}/workflows/${signal.workflow.id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-[9px] text-gray-400 dark:text-gray-500 hover:text-primary-500 dark:hover:text-primary-400 shrink-0 ml-auto"
              title="View workflow in Temporal UI"
            >
              workflow →
            </a>
          )}

          {/* Expand button for lazy loading */}
          {canExpandRemote && (
            <button
              onClick={() => onExpand(signal.queue_id, signal.id)}
              disabled={isLoading}
              className="text-[9px] text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 border border-dashed border-primary-300 dark:border-primary-700 rounded px-1.5 py-0.5 shrink-0 ml-auto disabled:opacity-50"
            >
              {isLoading ? '...' : '▸ expand'}
            </button>
          )}
        </div>
      </div>

      {/* Children */}
      {hasChildren && !isCollapsed
        ? node.children!.map((child, i) => (
            <TreeNode
              key={child.signal?.id ?? `child-${i}`}
              node={child}
              depth={depth + 1}
              collapsed={collapsed}
              onToggle={onToggle}
              onExpand={onExpand}
              loadingId={loadingId}
              expandedRemote={expandedRemote}
              temporalUIUrl={temporalUIUrl}
            />
          ))
        : null}
    </>
  )
}

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

function statusString(s: string): string {
  const l = s.toLowerCase()
  if (l.includes('completed') || l.includes('success')) return 'completed'
  if (l.includes('failed') || l.includes('error')) return 'failed'
  if (l.includes('running') || l.includes('executing') || l.includes('active') || l.includes('in_progress')) return 'running'
  if (l.includes('queued') || l.includes('pending')) return 'queued'
  return s || 'unknown'
}

function statusDotColor(s: string): string {
  const l = s.toLowerCase()
  if (l.includes('completed') || l.includes('success')) return 'text-green-500'
  if (l.includes('failed') || l.includes('error')) return 'text-red-500'
  if (l.includes('running') || l.includes('executing') || l.includes('active') || l.includes('in_progress')) return 'text-primary-500 animate-pulse'
  if (l.includes('queued') || l.includes('pending')) return 'text-amber-500'
  return 'text-gray-400 dark:text-gray-500'
}

function mergeChild(
  tree: SignalGraphNode,
  targetId: string,
  childGraph: SignalGraphNode
): SignalGraphNode {
  if (!tree.signal) return tree
  if (tree.signal.id === targetId) {
    return {
      ...tree,
      workflow_info: childGraph.workflow_info ?? tree.workflow_info,
      children: childGraph.children ?? tree.children ?? [],
    }
  }
  if (tree.children) {
    return {
      ...tree,
      children: tree.children.map(c => mergeChild(c, targetId, childGraph)),
    }
  }
  return tree
}
