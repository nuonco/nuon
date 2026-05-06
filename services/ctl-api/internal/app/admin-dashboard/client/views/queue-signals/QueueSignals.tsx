import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useSearchParams } from 'react-router'
import { getQueueSignalsGlobal, getQueueSignalTypeOptions } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { SignalLink } from '@/components/common/SignalLink'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const QueueSignals = () => {
  const [searchParams] = useSearchParams()
  const ownerID = searchParams.get('owner_id') || undefined
  const [search, setSearch] = useState('')
  const [signalType, setSignalType] = useState('')
  const [namespace, setNamespace] = useState('')
  const [enqueued, setEnqueued] = useState('')
  const [page, setPage] = useState(1)

  const { data: typeOptions } = useQuery({
    queryKey: ['queue-signal-type-options', namespace],
    queryFn: () => getQueueSignalTypeOptions(namespace || undefined),
  })

  const { data, isLoading, error } = useQuery({
    queryKey: ['queue-signals-global', search, signalType, namespace, enqueued, ownerID, page],
    queryFn: () => getQueueSignalsGlobal({
      search,
      signal_type: signalType || undefined,
      namespace: namespace || undefined,
      enqueued: enqueued || undefined,
      owner_id: ownerID,
      page,
    }),
    refetchInterval: 10000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signals'} />

  const signals = data?.signals || []
  const totalPages = data?.total_pages || 1
  const signalTypes = typeOptions?.signal_types || []

  return (
    <div>
      <h1 className="page-heading">Queue signals</h1>

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search by ID, owner, queue..." />
        </div>
        <select
          value={signalType}
          onChange={(e) => { setSignalType(e.target.value); setPage(1) }}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          <option value="">All types</option>
          {signalTypes.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
        <select
          value={enqueued}
          onChange={(e) => { setEnqueued(e.target.value); setPage(1) }}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300"
        >
          <option value="">All signals</option>
          <option value="false">Not enqueued</option>
          <option value="true">Enqueued</option>
        </select>
      </div>

      <div className="mt-4 table-card">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Type</th>
              <th>Owner</th>
              <th>Queue</th>
              <th>Status</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {signals.map((signal) => (
              <tr key={signal.id}>
                <td>
                  <SignalLink signalId={signal.id} queueId={signal.queue_id} />
                </td>
                <td><Badge>{signal.type}</Badge></td>
                <td className="text-gray-500 dark:text-gray-400">
                  <span className="font-mono text-xs">{truncateId(signal.owner_id)}</span>
                  <span className="ml-1 text-xs text-gray-400 dark:text-gray-500">({signal.owner_type})</span>
                </td>
                <td className="text-gray-500 dark:text-gray-400 font-mono text-xs">{truncateId(signal.queue_id)}</td>
                <td>
                  <Badge variant="status" status={String(signal.status?.status || signal.status)}>{String(signal.status?.status || signal.status)}</Badge>
                </td>
                <td className="text-gray-500 dark:text-gray-400">{formatDate(signal.created_at)}</td>
              </tr>
            ))}
            {signals.length === 0 && (
              <tr>
                <td colSpan={6} className="text-center text-gray-500 dark:text-gray-400 py-6">No signals found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
