import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getSignalCatalogDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const SignalCatalogDetail = () => {
  const { signalType } = useParams<{ signalType: string }>()

  const { data, isLoading, error } = useQuery({
    queryKey: ['signal-catalog-detail', signalType],
    queryFn: () => getSignalCatalogDetail(signalType!),
    enabled: !!signalType,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal type'} />
  if (!data) return null

  const { info, recent_signals = [] } = data

  return (
    <div className="space-y-6">
      <div>
        <h1 className="page-heading font-mono">{signalType}</h1>
        {info && (
          <div className="mt-2 flex flex-wrap gap-3 text-xs text-gray-500">
            {info.Namespace && <span>Namespace: <strong>{info.Namespace}</strong></span>}
            {info.Operation && <span>Operation: <strong>{info.Operation}</strong></span>}
          </div>
        )}
      </div>

      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Recent signals ({recent_signals.length})</h2>
        <div className="mt-2 table-card">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Owner</th>
                <th>Queue</th>
                <th>Status</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {recent_signals.map((signal) => (
                <tr key={signal.id}>
                  <td className="text-gray-500 font-mono text-xs">{truncateId(signal.id)}</td>
                  <td className="text-gray-500">
                    <span className="font-mono text-xs">{truncateId(signal.owner_id)}</span>
                    <span className="ml-1 text-xs text-gray-400">({signal.owner_type})</span>
                  </td>
                  <td className="text-gray-500 font-mono text-xs">{truncateId(signal.queue_id)}</td>
                  <td>
                    <Badge variant="status" status={String(signal.status?.status || signal.status)}>{String(signal.status?.status || signal.status)}</Badge>
                  </td>
                  <td className="text-gray-500">{formatDate(signal.created_at)}</td>
                </tr>
              ))}
              {recent_signals.length === 0 && (
                <tr>
                  <td colSpan={5} className="text-center text-gray-500 py-6">No recent signals</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
