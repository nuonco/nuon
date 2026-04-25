import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { getQueueSignalsGlobal, getQueueSignalTypeOptions } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const QueueSignals = () => {
  const [search, setSearch] = useState('')
  const [signalType, setSignalType] = useState('')
  const [page, setPage] = useState(1)

  const { data: typeOptions } = useQuery({
    queryKey: ['queue-signal-type-options'],
    queryFn: () => getQueueSignalTypeOptions(),
  })

  const { data, isLoading, error } = useQuery({
    queryKey: ['queue-signals-global', search, signalType, page],
    queryFn: () => getQueueSignalsGlobal({ search, signal_type: signalType || undefined, page }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signals'} />

  const { signals = [], total_pages = 1 } = data || {}
  const signalTypes = typeOptions?.signal_types || []

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Queue Signals</h1>

      <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search signals..." />
        </div>
        <select
          value={signalType}
          onChange={(e) => { setSignalType(e.target.value); setPage(1) }}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
        >
          <option value="">All types</option>
          {signalTypes.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Owner</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Queue</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {signals.map((signal) => (
              <tr key={signal.id}>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(signal.id)}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Badge>{signal.type}</Badge>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">
                  <span className="font-mono text-xs">{truncateId(signal.owner_id)}</span>
                  <span className="ml-1 text-xs text-gray-400">({signal.owner_type})</span>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(signal.queue_id)}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Badge variant="status" status={String(signal.status)}>{String(signal.status)}</Badge>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(signal.created_at)}</td>
              </tr>
            ))}
            {signals.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500">No signals found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
