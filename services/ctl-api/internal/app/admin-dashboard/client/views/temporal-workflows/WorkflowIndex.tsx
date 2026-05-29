import { useQuery } from '@tanstack/react-query'
import { useState, useCallback, useRef } from 'react'
import { Link } from 'react-router'
import { getTemporalWorkflowNamespaces, getBasePath } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { truncateId } from '@/utils/format'
import { DateTime } from 'luxon'

interface WorkflowEntry {
  workflow_id: string
  run_id: string
  workflow_type: string
  namespace: string
  status: string
  start_time: string
  history_length: number
  history_size_bytes: number
  can_count: number
  memo?: Record<string, string>
  is_queue: boolean
  queue_id?: string
  link?: string
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function age(startTime: string): string {
  if (!startTime) return '-'
  const dt = DateTime.fromISO(startTime)
  if (!dt.isValid) return '-'
  const diff = dt.diffNow().negate()
  if (diff.as('days') >= 1) return `${Math.floor(diff.as('days'))}d`
  if (diff.as('hours') >= 1) return `${Math.floor(diff.as('hours'))}h`
  return `${Math.floor(diff.as('minutes'))}m`
}

export const WorkflowIndex = () => {
  const [namespace, setNamespace] = useState('')
  const [workflows, setWorkflows] = useState<WorkflowEntry[]>([])
  const [indexing, setIndexing] = useState(false)
  const [summary, setSummary] = useState<{ total: number } | null>(null)
  const [sortBy, setSortBy] = useState<'history_length' | 'start_time' | 'workflow_type'>('history_length')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [filterType, setFilterType] = useState('')
  const abortRef = useRef<AbortController | null>(null)

  const { data: nsData } = useQuery({
    queryKey: ['temporal-workflow-namespaces'],
    queryFn: getTemporalWorkflowNamespaces,
  })

  const namespaces = nsData?.namespaces || []
  const temporalUIUrl = nsData?.temporal_ui_url || ''

  const startIndex = useCallback(async () => {
    if (!namespace || indexing) return

    // Reset state.
    setWorkflows([])
    setSummary(null)
    setIndexing(true)

    const controller = new AbortController()
    abortRef.current = controller

    try {
      const resp = await fetch(`${getBasePath()}/api/temporal-workflows/index?namespace=${encodeURIComponent(namespace)}`, {
        signal: controller.signal,
      })
      if (!resp.ok || !resp.body) {
        setIndexing(false)
        return
      }

      const reader = resp.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        const newEntries: WorkflowEntry[] = []
        for (const line of lines) {
          if (!line.trim()) continue
          try {
            const parsed = JSON.parse(line)
            if (parsed._type === 'summary') {
              setSummary({ total: parsed.total })
            } else {
              newEntries.push(parsed)
            }
          } catch { /* skip bad lines */ }
        }

        if (newEntries.length > 0) {
          setWorkflows(prev => [...prev, ...newEntries])
        }
      }
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        console.error('workflow index error', err)
      }
    } finally {
      setIndexing(false)
      abortRef.current = null
    }
  }, [namespace, indexing])

  const stopIndex = useCallback(() => {
    abortRef.current?.abort()
  }, [])

  // Sort and filter.
  const workflowTypes = [...new Set(workflows.map(w => w.workflow_type))].sort()
  const filtered = filterType ? workflows.filter(w => w.workflow_type === filterType) : workflows
  const sorted = [...filtered].sort((a, b) => {
    let cmp = 0
    if (sortBy === 'history_length') cmp = a.history_length - b.history_length
    else if (sortBy === 'start_time') cmp = a.start_time.localeCompare(b.start_time)
    else if (sortBy === 'workflow_type') cmp = a.workflow_type.localeCompare(b.workflow_type)
    return sortDir === 'desc' ? -cmp : cmp
  })

  const toggleSort = (col: typeof sortBy) => {
    if (sortBy === col) setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    else { setSortBy(col); setSortDir('desc') }
  }

  const sortIcon = (col: typeof sortBy) => sortBy === col ? (sortDir === 'desc' ? ' ▾' : ' ▴') : ''

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">Workflow Index</h1>

      {/* Controls */}
      <div className="flex flex-wrap items-center gap-3">
        <select
          value={namespace}
          onChange={(e) => setNamespace(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          <option value="">Select namespace...</option>
          {namespaces.map((ns) => (
            <option key={ns} value={ns}>{ns}</option>
          ))}
        </select>

        {!indexing ? (
          <button
            onClick={startIndex}
            disabled={!namespace}
            className="rounded-md bg-primary-600 dark:bg-primary-500 px-4 py-1.5 text-sm font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600 disabled:opacity-50"
          >
            Index Workflows
          </button>
        ) : (
          <button
            onClick={stopIndex}
            className="rounded-md bg-red-600 dark:bg-red-500 px-4 py-1.5 text-sm font-medium text-white hover:bg-red-700 dark:hover:bg-red-600"
          >
            Stop
          </button>
        )}

        {indexing && (
          <span className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <LoadingSpinner /> Loaded {workflows.length} workflows...
          </span>
        )}

        {summary && !indexing && (
          <span className="text-sm text-gray-500 dark:text-gray-400">
            {summary.total} running workflows indexed
          </span>
        )}
      </div>

      {/* Filters */}
      {workflows.length > 0 && (
        <div className="flex items-center gap-3">
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
          >
            <option value="">All types ({workflows.length})</option>
            {workflowTypes.map((t) => (
              <option key={t} value={t}>{t} ({workflows.filter(w => w.workflow_type === t).length})</option>
            ))}
          </select>
          <span className="text-xs text-gray-500 dark:text-gray-400">
            Showing {sorted.length} of {workflows.length}
          </span>
        </div>
      )}

      {/* Table */}
      {sorted.length > 0 && (
        <div className="table-card">
          <table>
            <thead>
              <tr>
                <th>Workflow ID</th>
                <th className="cursor-pointer select-none" onClick={() => toggleSort('workflow_type')}>
                  Type{sortIcon('workflow_type')}
                </th>
                <th>Status</th>
                <th className="cursor-pointer select-none" onClick={() => toggleSort('history_length')}>
                  Events{sortIcon('history_length')}
                </th>
                <th>Size</th>
                <th>CAN</th>
                <th className="cursor-pointer select-none" onClick={() => toggleSort('start_time')}>
                  Age{sortIcon('start_time')}
                </th>
                <th>Memo</th>
                <th>Link</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {sorted.map((wf) => (
                <tr key={wf.run_id}>
                  <td className="font-mono text-xs">
                    {temporalUIUrl ? (
                      <a
                        href={`${temporalUIUrl}/namespaces/${wf.namespace}/workflows/${wf.workflow_id}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
                      >
                        {truncateId(wf.workflow_id)}
                      </a>
                    ) : (
                      truncateId(wf.workflow_id)
                    )}
                  </td>
                  <td>
                    <Badge>{wf.workflow_type}</Badge>
                  </td>
                  <td>
                    <Badge variant="status" status={wf.status}>{wf.status}</Badge>
                  </td>
                  <td className="font-mono text-xs text-gray-500 dark:text-gray-400">
                    {wf.history_length.toLocaleString()}
                  </td>
                  <td className="font-mono text-xs text-gray-500 dark:text-gray-400">
                    {formatBytes(wf.history_size_bytes)}
                  </td>
                  <td className="font-mono text-xs text-gray-500 dark:text-gray-400">
                    {wf.can_count > 0 ? wf.can_count.toLocaleString() : '-'}
                  </td>
                  <td className="text-xs text-gray-500 dark:text-gray-400" title={wf.start_time}>
                    {age(wf.start_time)}
                  </td>
                  <td className="text-xs text-gray-500 dark:text-gray-400 max-w-[200px] truncate">
                    {wf.memo && Object.keys(wf.memo).length > 0
                      ? Object.entries(wf.memo).map(([k, v]) => `${k}=${v}`).join(', ')
                      : '-'}
                  </td>
                  <td>
                    {wf.link ? (
                      <Link
                        to={wf.link}
                        className="text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
                      >
                        {wf.is_queue ? 'Queue' : 'View'}
                      </Link>
                    ) : (
                      <span className="text-xs text-gray-400">-</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
