import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { bulkCancelWorkflows, getWorkflowFilterOptions, getWorkflows, type TBulkCancelResponse } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { ConfirmModal } from '@/components/common/ConfirmModal'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

const AWAITING_RETRY = 'failed-pending-retry'

const selectClass =
  'rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700'

const labelClass = 'flex flex-col gap-1 text-[11px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500'

function getStatus(status: any): string {
  if (!status) return ''
  if (typeof status === 'string') return status
  if (typeof status === 'object' && status.status) return String(status.status)
  return String(status)
}

// datetime-local yields local wall-clock; the BFF expects RFC3339
function toRFC3339(value: string): string | undefined {
  if (!value) return undefined
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return undefined
  return parsed.toISOString()
}

export const BulkCancelWorkflows = () => {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const [search, setSearch] = useState('')
  const [type, setType] = useState('')
  const [status, setStatus] = useState(AWAITING_RETRY)
  const [createdAfter, setCreatedAfter] = useState('')
  const [createdBefore, setCreatedBefore] = useState('')
  const [sort, setSort] = useState<'newest' | 'oldest'>('newest')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(20)

  const [selected, setSelected] = useState<string[]>([])
  const [confirming, setConfirming] = useState<'selected' | 'all' | null>(null)
  const [result, setResult] = useState<TBulkCancelResponse | null>(null)
  const [cancelError, setCancelError] = useState<string | null>(null)

  const { data: options } = useQuery({
    queryKey: ['workflow-filter-options'],
    queryFn: getWorkflowFilterOptions,
    staleTime: Infinity,
  })

  const workflowTypes = options?.types || []
  const cancelableStatuses = options?.cancelable_statuses || []
  const perPageOptions = options?.per_page_options || [20]

  const filters = useMemo(
    () => ({
      search: search || undefined,
      type: type || undefined,
      status: status || undefined,
      created_after: toRFC3339(createdAfter),
      created_before: toRFC3339(createdBefore),
    }),
    [search, type, status, createdAfter, createdBefore],
  )

  const resetPage = () => {
    setPage(1)
    setSelected([])
  }

  const { data, isLoading, error } = useQuery({
    queryKey: ['bulk-cancel-workflows', filters, sort, page, perPage],
    queryFn: () => getWorkflows({ ...filters, sort, page, per_page: perPage }),
    refetchInterval: 10_000,
  })

  const cancelMutation = useMutation({
    mutationFn: (body: { workflow_ids?: string[] }) => bulkCancelWorkflows({ ...filters, ...body }),
    onSuccess: (resp) => {
      setResult(resp)
      setCancelError(null)
      setSelected([])
      setConfirming(null)
      queryClient.invalidateQueries({ queryKey: ['bulk-cancel-workflows'] })
      if (resp.queue_id && resp.signal_id) {
        navigate(`/queues/${resp.queue_id}/signals/${resp.signal_id}`)
      }
    },
    onError: (err: any) => {
      setCancelError(err?.error || err?.message || 'Failed to start bulk cancel')
      setConfirming(null)
    },
  })

  const workflows = data?.workflows || []
  const total = data?.total ?? 0
  const totalPages = data?.total_pages || 1
  const pageIDs = workflows.map((wf: any) => wf.id)
  const allOnPageSelected = pageIDs.length > 0 && pageIDs.every((id: string) => selected.includes(id))

  const toggle = (id: string) =>
    setSelected((prev) => (prev.includes(id) ? prev.filter((v) => v !== id) : [...prev, id]))

  const toggleAllOnPage = () =>
    setSelected((prev) =>
      allOnPageSelected
        ? prev.filter((id) => !pageIDs.includes(id))
        : [...prev, ...pageIDs.filter((id: string) => !prev.includes(id))],
    )

  const confirmDescription =
    confirming === 'all'
      ? `This cancels all ${total} workflows matching the current filter. A background workflow cancels them one at a time; workflows that finished in the meantime are skipped.`
      : `This cancels the ${selected.length} selected workflow${selected.length === 1 ? '' : 's'}. A background workflow cancels them one at a time; workflows that finished in the meantime are skipped.`

  return (
    <div>
      <h1 className="page-heading">Bulk cancel workflows</h1>
      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Workflows stuck in a cancelable status, filterable by date range, type, and status.
      </p>

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); resetPage() }} placeholder="Search by ID or owner ID..." />
        </div>
        <label className={labelClass}>
          Status
          <select value={status} onChange={(e) => { setStatus(e.target.value); resetPage() }} className={selectClass}>
            {cancelableStatuses.map((s) => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>
        </label>
        <label className={labelClass}>
          Type
          <select value={type} onChange={(e) => { setType(e.target.value); resetPage() }} className={selectClass}>
            <option value="">All types</option>
            {workflowTypes.map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
        </label>
        <label className={labelClass}>
          Created after
          <input type="datetime-local" value={createdAfter} onChange={(e) => { setCreatedAfter(e.target.value); resetPage() }} className={selectClass} />
        </label>
        <label className={labelClass}>
          Created before
          <input type="datetime-local" value={createdBefore} onChange={(e) => { setCreatedBefore(e.target.value); resetPage() }} className={selectClass} />
        </label>
        <label className={labelClass}>
          Sort
          <select value={sort} onChange={(e) => { setSort(e.target.value as 'newest' | 'oldest'); resetPage() }} className={selectClass}>
            <option value="newest">Newest first</option>
            <option value="oldest">Oldest first</option>
          </select>
        </label>
        <label className={labelClass}>
          Per page
          <select value={perPage} onChange={(e) => { setPerPage(Number(e.target.value)); setPage(1) }} className={selectClass}>
            {perPageOptions.map((n) => (
              <option key={n} value={n}>{n}</option>
            ))}
          </select>
        </label>
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-3">
        <span className="text-sm text-gray-600 dark:text-gray-400">
          <span className="font-medium text-gray-900 dark:text-gray-100">{total}</span> matching
          {selected.length > 0 && <> &middot; {selected.length} selected</>}
        </span>
        <button
          onClick={() => setConfirming('selected')}
          disabled={selected.length === 0 || cancelMutation.isPending}
          className="rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-40 dark:bg-red-500 dark:hover:bg-red-600"
        >
          Cancel selected{selected.length > 0 ? ` (${selected.length})` : ''}
        </button>
        <button
          onClick={() => setConfirming('all')}
          disabled={total === 0 || cancelMutation.isPending}
          className="rounded-md px-3 py-1.5 text-sm font-medium text-red-700 ring-1 ring-inset ring-red-300 hover:bg-red-50 disabled:opacity-40 dark:text-red-400 dark:ring-red-800 dark:hover:bg-red-950"
        >
          Cancel all {total} matching
        </button>
      </div>

      {cancelError && <div className="mt-3"><ErrorMessage message={cancelError} /></div>}
      {result && (
        <div className="mt-3 rounded-md bg-gray-50 px-3 py-2 text-sm text-gray-700 dark:bg-gray-900 dark:text-gray-300">
          {result.message}
          {result.queue_id && result.signal_id && (
            <Link to={`/queues/${result.queue_id}/signals/${result.signal_id}`} className="ml-1 font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
              {result.signal_id}
            </Link>
          )}
        </div>
      )}

      {isLoading && <LoadingSpinner />}
      {error && <ErrorMessage message={(error as Error).message || 'Failed to load workflows'} />}

      {!isLoading && !error && (
        <div className="mt-4 table-card">
          <table>
            <thead>
              <tr>
                <th className="w-8">
                  <input type="checkbox" checked={allOnPageSelected} onChange={toggleAllOnPage} aria-label="Select all on page" />
                </th>
                <th>ID</th>
                <th>Type</th>
                <th>Owner</th>
                <th>Org</th>
                <th>Status</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {workflows.map((wf: any) => (
                <tr key={wf.id}>
                  <td>
                    <input type="checkbox" checked={selected.includes(wf.id)} onChange={() => toggle(wf.id)} aria-label={`Select ${wf.id}`} />
                  </td>
                  <td>
                    <Link to={`/workflows/${wf.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-mono text-xs">
                      {truncateId(wf.id)}
                    </Link>
                  </td>
                  <td className="font-mono text-xs text-gray-900 dark:text-gray-100">{wf.type}</td>
                  <td className="text-gray-500 dark:text-gray-400">
                    <Link to={`/installs/${wf.owner_id}`} className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                      {truncateId(wf.owner_id)}
                    </Link>
                    <span className="ml-1 text-[11px] text-gray-400 dark:text-gray-500">({wf.owner_type})</span>
                  </td>
                  <td className="text-gray-500 dark:text-gray-400">
                    <Link to={`/orgs/${wf.org_id}`} className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                      {truncateId(wf.org_id)}
                    </Link>
                  </td>
                  <td>
                    <Badge variant="status" status={getStatus(wf.status)}>{getStatus(wf.status) || '-'}</Badge>
                  </td>
                  <td className="text-gray-500 dark:text-gray-400">{formatDate(wf.created_at)}</td>
                </tr>
              ))}
              {workflows.length === 0 && (
                <tr>
                  <td colSpan={7} className="text-center text-gray-500 dark:text-gray-400 py-6">No workflows found</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />

      <ConfirmModal
        open={confirming !== null}
        title="Cancel workflows"
        description={confirmDescription}
        confirmLabel="Cancel workflows"
        confirmVariant="danger"
        isPending={cancelMutation.isPending}
        onCancel={() => setConfirming(null)}
        onConfirm={() => cancelMutation.mutate(confirming === 'selected' ? { workflow_ids: selected } : {})}
      />
    </div>
  )
}
