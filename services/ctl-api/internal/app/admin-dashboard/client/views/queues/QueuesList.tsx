import { useQuery, useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useSearchParams } from 'react-router'
import { getQueues, fullSweep, flushLostSignals } from '@/lib/admin-api'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { ConfirmModal } from '@/components/common/ConfirmModal'
import { formatDate, truncateId } from '@/utils/format'

export const QueuesList = () => {
  const [searchParams] = useSearchParams()
  const ownerID = searchParams.get('owner_id') || undefined
  const ownerType = searchParams.get('owner_type') || undefined
  const [search, setSearch] = useState('')
  const [name, setName] = useState('')
  const [namespace, setNamespace] = useState('')
  const [page, setPage] = useState(1)
  const [activeModal, setActiveModal] = useState<'flush' | null>(null)
  const [sweepResult, setSweepResult] = useState<{ enqueued: number; errors: number; duration_ms: number } | null>(null)
  const [flushResult, setFlushResult] = useState<{ flushed: number } | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['queues', search, name, namespace, ownerID, ownerType, page],
    queryFn: () => getQueues({
      search,
      name: name || undefined,
      namespace: namespace || undefined,
      owner_id: ownerID,
      owner_type: ownerType,
      page,
    }),
    refetchInterval: 20000,
  })

  const fullSweepMutation = useMutation({
    mutationFn: () => fullSweep(),
    onSuccess: (data) => { setSweepResult(data); setTimeout(() => setSweepResult(null), 10000) },
  })

  const flushMutation = useMutation({
    mutationFn: () => flushLostSignals(),
    onSuccess: (data) => { setActiveModal(null); setFlushResult(data); setTimeout(() => setFlushResult(null), 10000) },
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load queues'} />

  const { queues = [], total_pages = 1 } = data || {}

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">Queues</h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => fullSweepMutation.mutate()}
            disabled={fullSweepMutation.isPending}
            className="rounded-md bg-primary-600 dark:bg-primary-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600 disabled:opacity-50"
          >
            {fullSweepMutation.isPending ? 'Sweeping...' : 'Full Sweep'}
          </button>
          <button
            onClick={() => setActiveModal('flush')}
            disabled={flushMutation.isPending}
            className="rounded-md bg-red-600 dark:bg-red-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 dark:hover:bg-red-600 disabled:opacity-50"
          >
            {flushMutation.isPending ? 'Flushing...' : 'Flush Lost Signals'}
          </button>
        </div>
      </div>

      {sweepResult && (
        <div className="mt-2 rounded-md bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 p-3 text-sm text-green-800 dark:text-green-200">
          Full sweep complete: {sweepResult.enqueued} enqueued, {sweepResult.errors} errors, {sweepResult.duration_ms}ms
        </div>
      )}
      {flushResult && (
        <div className="mt-2 rounded-md bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 p-3 text-sm text-yellow-800 dark:text-yellow-200">
          Flushed {flushResult.flushed} lost signal(s)
        </div>
      )}
      {fullSweepMutation.isError && (
        <div className="mt-2 rounded-md bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-3 text-sm text-red-800 dark:text-red-200">
          Full sweep failed: {(fullSweepMutation.error as Error).message}
        </div>
      )}

      <ConfirmModal
        open={activeModal === 'flush'}
        title="Flush Lost Signals"
        description={'This will mark all unenqueued signals older than 1 hour as lost (error status) and soft-delete them.\n\nThese signals will no longer be retried by the sweep. This action cannot be undone.'}
        confirmLabel="Flush Lost Signals"
        confirmVariant="danger"
        onConfirm={() => flushMutation.mutate()}
        onCancel={() => setActiveModal(null)}
        isPending={flushMutation.isPending}
      />

      <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search queues..." />
        </div>
        <input
          type="text"
          value={name}
          onChange={(e) => { setName(e.target.value); setPage(1) }}
          placeholder="Filter by name..."
          className="block w-48 rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
        />
        <input
          type="text"
          value={namespace}
          onChange={(e) => { setNamespace(e.target.value); setPage(1) }}
          placeholder="Filter by namespace..."
          className="block w-48 rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
        />
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
          <thead className="">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Owner</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Emitters</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {queues.map((queue) => (
              <tr key={queue.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/queues/${queue.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-mono">
                    {truncateId(queue.id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{queue.name}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                  <span className="font-mono text-xs">{truncateId(queue.owner_id)}</span>
                  <span className="ml-1 text-xs text-gray-400 dark:text-gray-500">({queue.owner_type})</span>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{queue.emitters?.length ?? 0}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(queue.created_at)}</td>
              </tr>
            ))}
            {queues.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No queues found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
