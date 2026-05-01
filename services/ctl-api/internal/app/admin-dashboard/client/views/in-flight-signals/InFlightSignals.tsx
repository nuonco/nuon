import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getInFlightSignals } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { SignalLink } from '@/components/common/SignalLink'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

function getElapsed(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const updated = new Date(dateStr).getTime()
  const now = Date.now()
  const diffMs = now - updated
  if (diffMs < 0) return '-'
  const secs = Math.floor(diffMs / 1000)
  if (secs < 60) return `${secs}s`
  const mins = Math.floor(secs / 60)
  if (mins < 60) return `${mins}m ${secs % 60}s`
  const hours = Math.floor(mins / 60)
  return `${hours}h ${mins % 60}m`
}

export const InFlightSignals = () => {
  const [namespace, setNamespace] = useState('')
  const { data, isLoading, error } = useQuery({
    queryKey: ['in-flight-signals', namespace],
    queryFn: () => getInFlightSignals({ namespace: namespace || undefined }),
    refetchInterval: 5000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load in-flight signals'} />

  const signals = data?.signals || []
  const namespaces = data?.namespaces || []

  return (
    <div>
      <h1 className="page-heading">In-flight signals</h1>
      <p className="mt-1 text-sm text-gray-500">{signals.length} active signals (auto-refreshing)</p>

      <div className="mt-3">
        <select
          value={namespace}
          onChange={(e) => setNamespace(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300"
        >
          <option value="">All namespaces</option>
          {namespaces.map((n) => (
            <option key={n} value={n}>{n}</option>
          ))}
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
              <th>Elapsed</th>
              <th>Updated</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {signals.map((signal) => (
              <tr key={signal.id}>
                <td>
                  <SignalLink signalId={signal.id} queueId={signal.queue_id} />
                </td>
                <td className="whitespace-nowrap text-sm">
                  <Badge>{signal.type}</Badge>
                </td>
                <td className="whitespace-nowrap text-sm text-gray-500">
                  <span className="font-mono text-xs">{truncateId(signal.owner_id)}</span>
                  <span className="ml-1 text-xs text-gray-400">({signal.owner_type})</span>
                </td>
                <td className="whitespace-nowrap text-sm text-gray-500 font-mono">
                  <Link to={`/queues/${signal.queue_id}`} className="text-primary-600 hover:text-primary-700">
                    {truncateId(signal.queue_id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap text-sm">
                  <Badge variant="status" status={String(signal.status?.status || signal.status)}>{String(signal.status?.status || signal.status)}</Badge>
                </td>
                <td className="whitespace-nowrap text-sm text-gray-500 font-mono">
                  {getElapsed(signal.updated_at)}
                </td>
                <td className="whitespace-nowrap text-sm text-gray-500">{formatDate(signal.updated_at)}</td>
              </tr>
            ))}
            {signals.length === 0 && (
              <tr>
                <td colSpan={7} className="text-center text-gray-500 py-6">No in-flight signals</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
