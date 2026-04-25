import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getTemporalWorkerDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate } from '@/utils/format'

export const TemporalWorkerDetail = () => {
  const { namespace } = useParams<{ namespace: string }>()

  const { data, isLoading, error } = useQuery({
    queryKey: ['temporal-worker', namespace],
    queryFn: () => getTemporalWorkerDetail(namespace!),
    enabled: !!namespace,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load temporal worker'} />
  if (!data) return null

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">{data.namespace}</h1>
        <div className="mt-2 flex items-center gap-3 text-sm">
          <Badge variant="status" status={data.is_healthy ? 'healthy' : 'unhealthy'}>
            {data.is_healthy ? 'Healthy' : 'Unhealthy'}
          </Badge>
          <span className="text-gray-500">Task Queue: {data.task_queue}</span>
          <span className="text-gray-500">Total Pollers: {data.total_poller_count}</span>
        </div>
        {data.error && (
          <p className="mt-1 text-sm text-red-600">{data.error}</p>
        )}
      </div>

      {/* Workflow Pollers */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Workflow Pollers ({data.workflow_pollers?.length ?? 0})</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Identity</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Access</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rate/s</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(data.workflow_pollers || []).map((poller, i) => (
                <tr key={i}>
                  <td className="px-4 py-3 text-sm text-gray-900 break-all">{poller.identity}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(poller.last_access_time)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{poller.rate_per_second}</td>
                </tr>
              ))}
              {(!data.workflow_pollers || data.workflow_pollers.length === 0) && (
                <tr>
                  <td colSpan={3} className="px-4 py-8 text-center text-sm text-gray-500">No workflow pollers</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Activity Pollers */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Activity Pollers ({data.activity_pollers?.length ?? 0})</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Identity</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Access</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rate/s</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(data.activity_pollers || []).map((poller, i) => (
                <tr key={i}>
                  <td className="px-4 py-3 text-sm text-gray-900 break-all">{poller.identity}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(poller.last_access_time)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{poller.rate_per_second}</td>
                </tr>
              ))}
              {(!data.activity_pollers || data.activity_pollers.length === 0) && (
                <tr>
                  <td colSpan={3} className="px-4 py-8 text-center text-sm text-gray-500">No activity pollers</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Queue Stats */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Queue Stats</h2>
        <div className="mt-2 grid grid-cols-1 gap-4 sm:grid-cols-2">
          {data.workflow_stats && (
            <div className="rounded-md border border-gray-200 p-3">
              <h3 className="text-xs font-semibold text-gray-700">Workflow Queue</h3>
              <dl className="mt-2 grid grid-cols-2 gap-2 text-sm">
                <div>
                  <dt className="text-gray-500">Backlog Count</dt>
                  <dd className="text-gray-900">{data.workflow_stats.approximate_backlog_count}</dd>
                </div>
                <div>
                  <dt className="text-gray-500">Backlog Age</dt>
                  <dd className="text-gray-900">{data.workflow_stats.approximate_backlog_age}s</dd>
                </div>
                <div>
                  <dt className="text-gray-500">Add Rate</dt>
                  <dd className="text-gray-900">{data.workflow_stats.tasks_add_rate}/s</dd>
                </div>
                <div>
                  <dt className="text-gray-500">Dispatch Rate</dt>
                  <dd className="text-gray-900">{data.workflow_stats.tasks_dispatch_rate}/s</dd>
                </div>
              </dl>
            </div>
          )}
          {data.activity_stats && (
            <div className="rounded-md border border-gray-200 p-3">
              <h3 className="text-xs font-semibold text-gray-700">Activity Queue</h3>
              <dl className="mt-2 grid grid-cols-2 gap-2 text-sm">
                <div>
                  <dt className="text-gray-500">Backlog Count</dt>
                  <dd className="text-gray-900">{data.activity_stats.approximate_backlog_count}</dd>
                </div>
                <div>
                  <dt className="text-gray-500">Backlog Age</dt>
                  <dd className="text-gray-900">{data.activity_stats.approximate_backlog_age}s</dd>
                </div>
                <div>
                  <dt className="text-gray-500">Add Rate</dt>
                  <dd className="text-gray-900">{data.activity_stats.tasks_add_rate}/s</dd>
                </div>
                <div>
                  <dt className="text-gray-500">Dispatch Rate</dt>
                  <dd className="text-gray-900">{data.activity_stats.tasks_dispatch_rate}/s</dd>
                </div>
              </dl>
            </div>
          )}
          {!data.workflow_stats && !data.activity_stats && (
            <p className="text-sm text-gray-500">No queue stats available</p>
          )}
        </div>
      </div>
    </div>
  )
}
