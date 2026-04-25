import { useQuery } from '@tanstack/react-query'
import { getInFlightSignals } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const InFlightSignals = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['in-flight-signals'],
    queryFn: () => getInFlightSignals(),
    refetchInterval: 5000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load in-flight signals'} />

  const signals = data?.signals || []

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">In-flight signals</h1>
      <p className="mt-1 text-sm text-gray-500">{signals.length} active signals (auto-refreshing)</p>

      <div className="mt-4 overflow-x-auto rounded-lg border border-gray-200 bg-white">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Owner</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Queue</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Updated</th>
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
                  <Badge variant="status" status={String(signal.status?.status || signal.status)}>{String(signal.status?.status || signal.status)}</Badge>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(signal.updated_at)}</td>
              </tr>
            ))}
            {signals.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500">No in-flight signals</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
