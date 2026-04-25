import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getSignalCatalogDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const SignalCatalogDetail = () => {
  const { type } = useParams<{ type: string }>()

  const { data, isLoading, error } = useQuery({
    queryKey: ['signal-catalog-detail', type],
    queryFn: () => getSignalCatalogDetail(type!),
    enabled: !!type,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal type'} />
  if (!data) return null

  const { signal_type, recent_signals = [] } = data

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">{signal_type}</h1>
        <p className="mt-1 text-sm text-gray-500">{recent_signals.length} recent signals</p>
      </div>

      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Recent Signals</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Owner</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Queue</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {recent_signals.map((signal) => (
                <tr key={signal.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(signal.id)}</td>
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
              {recent_signals.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500">No recent signals</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
